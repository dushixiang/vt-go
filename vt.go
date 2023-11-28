package vt

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"unicode/utf8"
)

const (
	_BEL rune = 0x07 // Bell (Caret = ^G, C = \a)
	_BS  rune = 0x08 // Backspace (Caret = ^H, C = \b)
	_HT  rune = 0x09 // Position to the next character tab stop.(Caret = ^I, C = \t)
	_LF  rune = 0x0a // LF Line Feed (Caret = ^J, C = \n)
	_VT  rune = 0x0b // Position the form at the next line tab stop.(Caret = ^K, C = \v)
	_CR  rune = 0x0d // Carriage Return (Caret = ^M, C = \r)

	_ESC rune = 0x1b // Escape (Caret = ^[, C = \e)
	_DEL rune = 0x7f // Delete (Caret = ^?)

	_ST rune = 0x9c // String Terminator

	space rune = 0x20 // 空格
)

type inputHandler func(params []rune) error

type VirtualTerminal interface {
	Advance(p []byte) (int, error)
	Parse() []string
	Reset()
	Empty() bool
	Bytes() []byte
}

type Opts struct {
	Logger *log.Logger
}

func New(realtime bool) VirtualTerminal {
	vt := virtualTerminal{
		realtime:      realtime,
		inputHandlers: make(map[byte]inputHandler),
		rowList:       make([]*row, 0),
		rows:          0,
		logger:        nil,
	}
	vt.initCsiHandler()
	return &vt
}

type virtualTerminal struct {
	realtime   bool
	empty      atomic.Bool
	buffer     bytes.Buffer
	bufferLock sync.Mutex

	rowList []*row // 行数据
	rows    int    // 行数量

	inputHandlers map[byte]inputHandler
	insertMode    bool // 暂时没啥用
	logger        *log.Logger
}

func (vt *virtualTerminal) Empty() bool {
	return vt.empty.Load()
}

func (vt *virtualTerminal) addCsiHandler(b byte, handler inputHandler) {
	vt.inputHandlers[b] = handler
}

func (vt *virtualTerminal) getCurrentRow() *row {
	if len(vt.rowList) == 0 {
		vt.rowList = append(vt.rowList, vt.newRow())
		vt.rows = 1
	}

	if len(vt.rowList) < vt.rows {
		count := vt.rows - len(vt.rowList)
		for i := 0; i < count; i++ {
			vt.rowList = append(vt.rowList, vt.newRow())
		}
	}

	index := vt.rows - 1
	if index < 0 {
		index = 0
	}

	return vt.rowList[index]
}

func (vt *virtualTerminal) newRow() *row {
	return &row{
		data:  make([]rune, 0),
		index: 0,
	}
}

// https://zh.wikipedia.org/zh/ANSI%E8%BD%AC%E4%B9%89%E5%BA%8F%E5%88%97
func (vt *virtualTerminal) handleSequence(inputs []byte) []byte {
	code, size := utf8.DecodeRune(inputs)
	inputs = inputs[size:]
	switch code {
	case '[': // CSI - 控制序列导入器（Control Sequence Introducer）
		inputs = vt.handleCSISequence(inputs)
	case ']': // OSC – 操作系统命令（Operating System Command）
		inputs = vt.handleOSCSequence(inputs)
	}
	return inputs
}

func (vt *virtualTerminal) handleCSISequence(p []byte) []byte {
	index := bytes.IndexFunc(p, func(r rune) bool {
		return isCSISequence(r)
	})
	if index > -1 {
		b := p[index]
		handler, ok := vt.inputHandlers[b]
		if ok {
			params := []rune(string(p[:index]))
			if err := handler(params); err != nil {
				vt.log(fmt.Sprintf("handle csi sequence err %v", err.Error()))
			}
		} else {
			vt.log(fmt.Sprintf("no match input handler for %q %v", b, b))
		}
	}
	return p[index+1:]
}

// 启动操作系统使用的控制字符串。OSC序列与CSI序列相似，但不限于整数参数。
// 通常，这些控制序列由ST终止[12]:8.3.89。
// 在xterm中，它们也可能被BEL终止[13]。
// 例如，在xterm中，窗口标题可以这样设置：OSC 0;this is the window title _BEL。
func (vt *virtualTerminal) handleOSCSequence(p []byte) []byte {
	if index := bytes.IndexRune(p, _ST); index >= 0 {
		return p[index+1:]
	}
	if index := bytes.IndexRune(p, _BEL); index >= 0 {
		return p[index+1:]
	}
	return p
}

// https://zh.wikipedia.org/zh/C0%E4%B8%8EC1%E6%8E%A7%E5%88%B6%E5%AD%97%E7%AC%A6
func (vt *virtualTerminal) handleC0Sequence(code rune) {
	switch code {
	case _BEL: // \a 发出可听见的噪音。
	case _BS: // \b 将光标向左移动一个字符
		vt.moveBackward(1)
	case _HT: // \t 定位到下一个制表位。
		// TODO
	case _LF: // \n 将光标移动到下一行,但不改变所在的列的位置
		vt.moveDown(1)
	case _VT: // \v 定位到下一行的制表位。
		// TODO
	case _CR: // \r 将光标移动到当前行的最左边。
		vt.moveTo(0, vt.rows)
	case _DEL: // 最初用于穿孔纸带上删除一个字符。因为任何位置的字符都可以被全部穿孔（全1）。VT100兼容终端，按键⌫产生这个字符，常称为backspace，但不对应于PC键盘的delete key。
		// TODO
	}
}

func (vt *virtualTerminal) log(v ...interface{}) {
	if vt.logger != nil {
		log.Println(v)
	}
}

func (vt *virtualTerminal) getNumberOrDefault(params []rune, index, _default int) int {
	// 下标检查
	if len(params)-1 < index {
		return _default
	}
	n, err := strconv.Atoi(string(params[index]))
	if err != nil {
		n = _default
	}
	return n
}

func (vt *virtualTerminal) getNumberOrDefaultOfBytes(params []byte, index, _default int) int {
	// 下标检查
	if len(params)-1 < index {
		return _default
	}
	n, err := strconv.Atoi(string(params[index]))
	if err != nil {
		n = _default
	}
	return n
}

func (vt *virtualTerminal) appendCharacter(code rune) {
	row := vt.getCurrentRow()
	row.append(code)
}

func (vt *virtualTerminal) Advance(p []byte) (int, error) {
	if vt.realtime {
		inputs := make([]byte, len(p))
		copy(inputs, p)
		vt.advance(inputs)
		return len(p), nil
	} else {
		vt.bufferLock.Lock()
		defer vt.bufferLock.Unlock()
		vt.empty.Store(false)
		return vt.buffer.Write(p)
	}
}

func (vt *virtualTerminal) advance(inputs []byte) {
	for len(inputs) > 0 {
		code, size := utf8.DecodeRune(inputs)
		inputs = inputs[size:]
		if _ESC == code {
			inputs = vt.handleSequence(inputs)
			continue
		}
		if isC0Sequence(code) {
			vt.handleC0Sequence(code)
		} else {
			vt.appendCharacter(code)
		}
	}
}

func (vt *virtualTerminal) Parse() []string {
	if !vt.realtime {
		inputs := vt.Bytes()
		toString := base64.StdEncoding.EncodeToString(inputs)
		fmt.Printf("++++++++++++++++++++++++++++++++")
		fmt.Printf("%v\n", toString)
		fmt.Printf("%v\n", len(inputs))
		fmt.Printf("++++++++++++++++++++++++++++++++")
		vt.advance(inputs)
	}
	var result []string
	for i := range vt.rowList {
		line := vt.rowList[i].String()
		result = append(result, line)
	}
	return result
}

func (vt *virtualTerminal) Reset() {
	vt.bufferLock.Lock()
	defer vt.bufferLock.Unlock()
	_ = vt.eraseAll()
	vt.buffer.Reset()
	vt.empty.Store(true)
}

func (vt *virtualTerminal) Bytes() []byte {
	vt.bufferLock.Lock()
	defer vt.bufferLock.Unlock()
	b := vt.buffer.Bytes()
	nb := make([]byte, len(b))
	copy(nb, b)
	return nb
}
