package deelfer

import "bytes"

type Deelfer struct {
	dict map[rune]rune
}

var ArE = []rune("qwertyuiop[]asdfghjkl;'\\zxcvbnm,./QWERTYUIOP{}ASDFGHJKL:\"|ZXCVBNM<>?")
var ArR = []rune("йцукенгшщзхъфывапролджэ\\ячсмитьбю.ЙЦУКЕНГШЩЗХЪФЫВАПРОЛДЖЭ/ЯЧСМИТЬБЮ,")

func New() *Deelfer {
	d := &Deelfer{
		dict: map[rune]rune{},
	}

	for i, c := range ArE {
		d.dict[c] = ArR[i]
		d.dict[ArR[i]] = c
	}

	return d
}

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
