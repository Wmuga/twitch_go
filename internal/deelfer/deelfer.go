package deelfer

import "bytes"

// Switches layout eng symbols <-> ru symbols.
// Create new instance with NewDeelfer()
type Deelfer struct {
	dict map[rune]rune
}

// English symbols
var ArE = []rune("qwertyuiop[]asdfghjkl;'\\zxcvbnm,./QWERTYUIOP{}ASDFGHJKL:\"|ZXCVBNM<>?")

// Russians symbols
var ArR = []rune("йцукенгшщзхъфывапролджэ\\ячсмитьбю.ЙЦУКЕНГШЩЗХЪФЫВАПРОЛДЖЭ/ЯЧСМИТЬБЮ,")

// Creates new instance of Deelfer
func NewDeelfer() *Deelfer {
	d := &Deelfer{
		dict: map[rune]rune{},
	}

	for i, c := range ArE {
		d.dict[c] = ArR[i]
		d.dict[ArR[i]] = c
	}

	return d
}

// Switches layout
func (d *Deelfer) Translate(msg string) string {
	var buf bytes.Buffer
	for _, c := range msg {
		if t, ex := d.dict[c]; ex {
			buf.WriteRune(t) //nolint:errcheck
			continue
		}
		buf.WriteRune(c) //nolint:errcheck
	}
	return buf.String()
}
