package vt

func (vt *virtualTerminal) resetCursor() {
	vt.getCurrentRow().setIndex(0)
	vt.rows = 0
}

func (vt *virtualTerminal) moveTo(col, row int) {
	vt.setCol(col)
	vt.setRow(row)
}

func (vt *virtualTerminal) setRow(row int) {
	vt.rows = row
}

func (vt *virtualTerminal) setCol(col int) {
	vt.getCurrentRow().setIndex(col)
}

func (vt *virtualTerminal) moveUp(ps int) {
	vt.rows -= ps
	if vt.rows < 0 {
		vt.rows = 0
	}
}

func (vt *virtualTerminal) moveDown(ps int) {
	vt.rows += ps
}

func (vt *virtualTerminal) moveBackward(ps int) {
	index := vt.getCurrentRow().index
	index -= ps
	if index < 0 {
		index = 0
	}
	vt.setCol(index)
}

func (vt *virtualTerminal) moveForward(ps int) {
	index := vt.getCurrentRow().index + ps
	vt.setCol(index)
}

func (vt *virtualTerminal) move(col int, row int) {
	newCol := vt.getCurrentRow().index + col
	newRow := vt.rows + row
	vt.moveTo(newCol, newRow)
}
