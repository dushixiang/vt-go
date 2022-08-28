package vt

type Row struct {
	data  []rune // 当前行
	index int
}

func (r *Row) setIndex(index int) {
	r.index = index
}

// 添加新输入的字符
func (r *Row) append(code rune) {
	if r.index < len(r.data) {
		r.data[r.index] = code
	} else {
		r.data = append(r.data, code)
	}
	r.index++
}

// 向下标位置插入字符
func (r *Row) insert(code ...rune) {
	for _, c := range code {
		r.data = insert(r.data, r.index, c)
	}
}

// 从下标位置删除N个字符
func (r *Row) delete(ps int) {
	r.data = remove(r.data, r.index, ps)
}

// 删除当前光标所在位置右侧的字符
func (r *Row) eraseRight() {
	if r.index < len(r.data) {
		r.data = r.data[:r.index]
	}
}

// 删除当前光标所在位置左侧的字符
func (r *Row) eraseLeft() {
	r.data = r.data[r.index:]
}

func (r *Row) String() string {
	return string(r.data)
}
