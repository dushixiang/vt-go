package vt

func (vt *virtualTerminal) initCsiHandler() {
	vt.addCsiHandler('@', vt.insertChar)
	vt.addCsiHandler('A', vt.cursorUp)
	vt.addCsiHandler('B', vt.cursorDown)
	vt.addCsiHandler('C', vt.cursorForward)
	vt.addCsiHandler('D', vt.cursorBackward)
	vt.addCsiHandler('E', vt.cursorNextLine)
	vt.addCsiHandler('F', vt.cursorPrecedingLine)
	vt.addCsiHandler('G', vt.cursorCharAbsolute)
	vt.addCsiHandler('H', vt.cursorPosition)
	vt.addCsiHandler('J', vt.eraseInDisplay)
	vt.addCsiHandler('K', vt.eraseInLine)
	vt.addCsiHandler('P', vt.deleteChars)
	vt.addCsiHandler('X', vt.eraseChars)
	vt.addCsiHandler('`', vt.charPosAbsolute)
	vt.addCsiHandler('a', vt.hPositionRelative)
	vt.addCsiHandler('d', vt.linePosAbsolute)
	vt.addCsiHandler('e', vt.vPositionRelative)
	vt.addCsiHandler('f', vt.hVPosition)
	vt.addCsiHandler('h', vt.setMode)
	vt.addCsiHandler('l', vt.resetMode)
	vt.addCsiHandler('m', vt.charAttributes)
	vt.addCsiHandler('r', vt.setScrollRegion)
}

func (vt *virtualTerminal) cursorChange(params []rune, action func(ps int)) {
	ps := vt.getNumberOrDefault(params, 0, 1)
	action(ps)
}

// insert Ps (Blank) Character(s) (default = 1) (ICH).
func (vt *virtualTerminal) insertChar(params []rune) error {
	row := vt.getCurrentRow()
	ps := vt.getNumberOrDefault(params, 0, 1)
	for i := 0; i < ps; i++ {
		row.insert(space)
	}
	return nil
}

// 光标向指定的方向移动{n（默认1）格。如果光标已在屏幕边缘，则无效。
func (vt *virtualTerminal) cursorUp(params []rune) error {
	vt.cursorChange(params, vt.moveUp)
	return nil
}

// 光标向指定的方向移动{n（默认1）格。如果光标已在屏幕边缘，则无效。
func (vt *virtualTerminal) cursorDown(params []rune) error {
	vt.cursorChange(params, vt.moveDown)
	return nil
}

// 光标向指定的方向移动{n（默认1）格。如果光标已在屏幕边缘，则无效。
func (vt *virtualTerminal) cursorForward(params []rune) error {
	vt.cursorChange(params, vt.moveForward)
	return nil
}

// 光标向指定的方向移动{n（默认1）格。如果光标已在屏幕边缘，则无效。
func (vt *virtualTerminal) cursorBackward(params []rune) error {
	vt.cursorChange(params, vt.moveBackward)
	return nil
}

// 光标移动到下面第n（默认1）行的开头。
func (vt *virtualTerminal) cursorNextLine(params []rune) error {
	return vt.cursorDown(params)
}

// 光标移动到上面第n（默认1）行的开头。
func (vt *virtualTerminal) cursorPrecedingLine(params []rune) error {
	return vt.cursorUp(params)
}

// 光标移动到第n（默认1）列。
func (vt *virtualTerminal) cursorCharAbsolute(params []rune) error {
	vt.cursorChange(params, func(ps int) {
		vt.moveTo(ps, vt.rows)
	})
	return nil
}

// 光标移动到第n行、第m列。值从1开始，且默认为1（左上角）。
// 例如CSI ;5H和CSI 1;5H含义相同；CSI 17;H、CSI 17H和CSI 17;1H三者含义相同。
func (vt *virtualTerminal) cursorPosition(params []rune) error {
	if len(params) >= 2 {
		row := vt.getNumberOrDefault(params, 0, 1)
		col := vt.getNumberOrDefault(params, 1, 1)
		vt.moveTo(col, row)
	} else {
		vt.resetCursor()
	}
	return nil
}

// 清除屏幕的部分区域。如果n是0（或缺失），则清除从光标位置到屏幕末尾的部分。
// 如果n是1，则清除从光标位置到屏幕开头的部分。
// 如果n是2，则清除整个屏幕（在DOS ANSI.SYS中，光标还会向左上方移动）。
// 如果n是3，则清除整个屏幕，并删除回滚缓存区中的所有行（这个特性是xterm添加的，其他终端应用程序也支持）。
func (vt *virtualTerminal) eraseInDisplay(params []rune) error {
	ps := vt.getNumberOrDefault(params, 0, 0)
	switch ps {
	case 0:
		return vt.eraseBelow()
	case 1:
		return vt.eraseAbove()
	case 2:
		return vt.eraseAll()
	case 3:
		// 忽略 Erase Saved Lines (xterm)
	}
	return nil
}

// 清除行内的部分区域。
// 如果n是0（或缺失），清除从光标位置到该行末尾的部分。
// 如果n是1，清除从光标位置到该行开头的部分。
// 如果n是2，清除整行。光标位置不变。
func (vt *virtualTerminal) eraseInLine(params []rune) error {
	ps := vt.getNumberOrDefault(params, 0, 0)
	switch ps {
	case 0:
		return vt.eraseRight()
	case 1:
		return vt.eraseLeft()
	case 2:
		return vt.eraseAll()
	}
	return nil
}

// Delete Ps Character(s) (default = 1) (DCH).
func (vt *virtualTerminal) deleteChars(params []rune) error {
	ps := vt.getNumberOrDefault(params, 0, 1)
	row := vt.getCurrentRow()
	row.delete(ps)
	return nil
}

// Erase Ps Character(s) (default = 1) (ECH).
func (vt *virtualTerminal) eraseChars(params []rune) error {
	return vt.deleteChars(params)
}

// Character Position Absolute  [column] (default = [rows,1])
func (vt *virtualTerminal) charPosAbsolute(params []rune) error {
	ps := vt.getNumberOrDefault(params, 0, 1) - 1
	vt.moveTo(ps, vt.rows)
	return nil
}

// Character Position Relative  [columns] (default = [rows,col+1])
func (vt *virtualTerminal) hPositionRelative(params []rune) error {
	ps := vt.getNumberOrDefault(params, 0, 1)
	vt.move(ps, 0)
	return nil
}

// 行定位绝对[ROW]（default = [1，列]）（VPA）。
func (vt *virtualTerminal) linePosAbsolute(params []rune) error {
	ps := vt.getNumberOrDefault(params, 0, 1) - 1
	vt.setRow(ps)
	return nil
}

// Line Position Relative  [rowList] (default = [rows+1,column])
func (vt *virtualTerminal) vPositionRelative(params []rune) error {
	ps := vt.getNumberOrDefault(params, 0, 1)
	vt.moveTo(0, ps)
	return nil
}

// Horizontal and Vertical Position [rows;column] (default = [1,1]) (HVP).
func (vt *virtualTerminal) hVPosition(params []rune) error {
	return vt.cursorPosition(params)
}

/**
 * CSI Pm h  Set Mode (SM).
 *     Ps = 2  -> Keyboard Action Mode (AM).
 *     Ps = 4  -> insert Mode (IRM).
 *     Ps = 1 2  -> Send/receive (SRM).
 *     Ps = 2 0  -> Automatic Newline (LNM).
 *
 * @virtualTerminal: #P[Only IRM is supported.]    CSI SM    "Set Mode"  "CSI Pm h"  "Set various terminal modes."
 * Supported param values by SM:
 *
 * | Param | Action                                 | Support |
 * | ----- | -------------------------------------- | ------- |
 * | 2     | Keyboard Action Mode (KAM). Always on. | #N      |
 * | 4     | insert Mode (IRM).                     | #Y      |
 * | 12    | Send/receive (SRM). Always off.        | #N      |
 * | 20    | Automatic Newline (LNM). Always off.   | #N      |
 */
func (vt *virtualTerminal) setMode(params []rune) error {
	for _, param := range params {
		ps := vt.getNumberOrDefault([]rune{param}, 0, 0)
		if ps == 4 {
			vt.insertMode = true
			break
		}
	}
	return nil
}

func (vt *virtualTerminal) resetMode(params []rune) error {
	for _, param := range params {
		ps := vt.getNumberOrDefault([]rune{param}, 0, 0)
		if ps == 4 {
			vt.insertMode = false
			break
		}
	}
	return nil
}

// Set Scrolling Region [top;bottom] (default = full size of window) (DECSTBM), VT100.
func (vt *virtualTerminal) setScrollRegion(params []rune) error {
	top := vt.getNumberOrDefault(params, 0, 1)
	bottom := vt.getNumberOrDefault(params, 1, 0)
	if len(params) < 2 || bottom > len(vt.rowList) || bottom == 0 {
		bottom = len(vt.rowList)
	}
	if bottom > top {
		vt.moveTo(0, 0)
	}
	return nil
}

func (vt *virtualTerminal) eraseBelow() error {
	if len(vt.rowList) > vt.rows {
		vt.rowList = vt.rowList[:vt.rows]
	}
	return nil
}

func (vt *virtualTerminal) eraseAbove() error {
	vt.rowList = vt.rowList[vt.rows-1:]
	return nil
}

func (vt *virtualTerminal) eraseAll() error {
	vt.rowList = nil
	vt.resetCursor()
	return nil
}

func (vt *virtualTerminal) eraseRight() error {
	row := vt.getCurrentRow()
	row.eraseRight()
	return nil
}

func (vt *virtualTerminal) eraseLeft() error {
	row := vt.getCurrentRow()
	row.eraseLeft()
	return nil
}

func (vt *virtualTerminal) charAttributes(params []rune) error {
	// 忽略字符的颜色属性
	return nil
}
