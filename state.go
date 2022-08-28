package vt

func (vt *VirtualTerminal) resetCursor() {
	vt.getCurrentRow().setIndex(0)
	vt.rows = 0
}

func (vt *VirtualTerminal) moveTo(col, row int) {
	vt.setCol(col)
	vt.setRow(row)
}

func (vt *VirtualTerminal) setRow(row int) {
	vt.rows = row
}

func (vt *VirtualTerminal) setCol(col int) {
	vt.getCurrentRow().setIndex(col)
}

func (vt *VirtualTerminal) moveUp(ps int) {
	vt.rows -= ps
	if vt.rows < 0 {
		vt.rows = 0
	}
}

func (vt *VirtualTerminal) moveDown(ps int) {
	vt.rows += ps
}

func (vt *VirtualTerminal) moveBackward(ps int) {
	index := vt.getCurrentRow().index
	index -= ps
	if index < 0 {
		index = 0
	}
	vt.setCol(index)
}

func (vt *VirtualTerminal) moveForward(ps int) {
	index := vt.getCurrentRow().index + ps
	vt.setCol(index)
}

func (vt *VirtualTerminal) move(col int, row int) {
	newCol := vt.getCurrentRow().index + col
	newRow := vt.rows + row
	vt.moveTo(newCol, newRow)
}
