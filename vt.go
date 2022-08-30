package vt

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"unicode/utf8"
)

const (
	BEL rune = 0x07 // Bell (Caret = ^G, C = \a)
	BS  rune = 0x08 // Backspace (Caret = ^H, C = \b)
	HT  rune = 0x09 // Position to the next character tab stop.(Caret = ^I, C = \t)
	LF  rune = 0x0a // LF Line Feed (Caret = ^J, C = \n)
	VT  rune = 0x0b // Position the form at the next line tab stop.(Caret = ^K, C = \v)
	CR  rune = 0x0d // Carriage Return (Caret = ^M, C = \r)

	ESC rune = 0x1b // Escape (Caret = ^[, C = \e)
	DEL rune = 0x7f // Delete (Caret = ^?)

	ST rune = 0x9c // String Terminator

	space rune = 0x20 // 空格
)

type inputHandler func(params []rune) error

type VirtualTerminal struct {
	buffer  bytes.Buffer
	rowList []*Row // 行数据
	rows    int    // 行数量

	inputHandlers map[byte]inputHandler
	insertMode    bool // 暂时没啥用
	logger        *log.Logger
}

type NewVirtualTerminalOpts struct {
	Logger *log.Logger
}

func New() *VirtualTerminal {
	return NewWithOpts(NewVirtualTerminalOpts{})
}

func NewWithOpts(opts NewVirtualTerminalOpts) *VirtualTerminal {
	vt := &VirtualTerminal{
		inputHandlers: make(map[byte]inputHandler),
		rowList:       make([]*Row, 0),
		rows:          0,
		logger:        opts.Logger,
	}
	vt.initCsiHandler()
	return vt
}

func (vt *VirtualTerminal) addCsiHandler(b byte, handler inputHandler) {
	vt.inputHandlers[b] = handler
}

func (vt *VirtualTerminal) getCurrentRow() *Row {
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

func (vt *VirtualTerminal) newRow() *Row {
	return &Row{
		data:  make([]rune, 0),
		index: 0,
	}
}

// https://zh.wikipedia.org/zh/ANSI%E8%BD%AC%E4%B9%89%E5%BA%8F%E5%88%97
func (vt *VirtualTerminal) handleSequence(inputs []byte) []byte {
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

func (vt *VirtualTerminal) handleCSISequence(p []byte) []byte {
	index := bytes.IndexFunc(p, func(r rune) bool {
		return isCSISequence(r)
	})
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
	return p[index+1:]
}

// 启动操作系统使用的控制字符串。OSC序列与CSI序列相似，但不限于整数参数。
// 通常，这些控制序列由ST终止[12]:8.3.89。
// 在xterm中，它们也可能被BEL终止[13]。
// 例如，在xterm中，窗口标题可以这样设置：OSC 0;this is the window title BEL。
func (vt *VirtualTerminal) handleOSCSequence(p []byte) []byte {
	if index := bytes.IndexRune(p, ST); index >= 0 {
		return p[index+1:]
	}
	if index := bytes.IndexRune(p, BEL); index >= 0 {
		return p[index+1:]
	}
	return p
}

// https://zh.wikipedia.org/zh/C0%E4%B8%8EC1%E6%8E%A7%E5%88%B6%E5%AD%97%E7%AC%A6
func (vt *VirtualTerminal) handleC0Sequence(code rune) {
	switch code {
	case BEL: // \a 发出可听见的噪音。
	case BS: // \b 将光标向左移动一个字符
		vt.moveBackward(1)
	case HT: // \t 定位到下一个制表位。
		// TODO
	case LF: // \n 将光标移动到下一行,但不改变所在的列的位置
		vt.moveDown(1)
	case VT: // \v 定位到下一行的制表位。
		// TODO
	case CR: // \r 将光标移动到当前行的最左边。
		vt.moveTo(0, vt.rows)
	case DEL: // 最初用于穿孔纸带上删除一个字符。因为任何位置的字符都可以被全部穿孔（全1）。VT100兼容终端，按键⌫产生这个字符，常称为backspace，但不对应于PC键盘的delete key。
		// TODO
	}
}

func (vt *VirtualTerminal) log(v ...any) {
	if vt.logger != nil {
		log.Println(v)
	}
}

func (vt *VirtualTerminal) getNumberOrDefault(params []rune, index, _default int) int {
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

func (vt *VirtualTerminal) getNumberOrDefaultOfBytes(params []byte, index, _default int) int {
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

func (vt *VirtualTerminal) appendCharacter(code rune) {
	row := vt.getCurrentRow()
	row.append(code)
}

func (vt *VirtualTerminal) Advance(p []byte) (int, error) {
	return vt.buffer.Write(p)
}

func (vt *VirtualTerminal) Parse() {
	vt.moveDown(1)
	inputs := vt.buffer.Bytes()
	vt.buffer.Reset()
	for len(inputs) > 0 {
		code, size := utf8.DecodeRune(inputs)
		inputs = inputs[size:]
		if ESC == code {
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

func (vt *VirtualTerminal) Reset() {
	vt.resetCursor()
	vt.buffer.Reset()
	vt.rowList = nil
}

func (vt *VirtualTerminal) Result() []string {
	result := make([]string, len(vt.rowList))
	for i := range vt.rowList {
		line := vt.rowList[i].String()
		result[i] = line
	}
	return result
}
