package vt

import (
	"fmt"
	"testing"
	"unicode/utf8"
)

func TestParser(t *testing.T) {
	var tests = []struct {
		in  string
		out []string
	}{
		{
			"\r(reverse-i-search)`': \x1b[K\b\b\bp': ps -a\b\b\b\b\b\r\x1b[11@[root@FAT00400000 koko-allinone]#\x1b[C\x1b[C\x1b[C\x1b[C\x1b[C\x1b[C",
			[]string{"[root@FAT00400000 koko-allinone]# ps -a"},
		},
	}

	for _, test := range tests {
		terminal := New()
		terminal.Advance([]byte(test.in))
		out := terminal.Output()
		if !testEq(out, test.out) {
			t.Errorf("expected %#v got %#v", test.out, out)
		}
	}
}

func testEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestInput(t *testing.T) {
	var in = "\n(reverse-i-search)`': \u001B[K\b\b\bp': ps -a\b\b\b\b\b\n\u001B[11@[root@FAT00400000 koko-allinone]#\u001B[C\u001B[C\u001B[C\u001B[C\u001B[C\u001B[C"

	var inputs = []byte(in)

	for len(inputs) > 0 {
		code, size := utf8.DecodeRune(inputs)
		inputs = inputs[size:]
		fmt.Print(string(code))
		//time.Sleep(time.Millisecond * 10)
		//if _ESC == code {
		//	inputs = vt.handleSequence(inputs)
		//	continue
		//}
		//if isC0Sequence(code) {
		//	vt.handleC0Sequence(code)
		//} else {
		//	vt.appendCharacter(code)
		//}
	}
	fmt.Println()
}
