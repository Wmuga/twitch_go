package tools

import "strconv"

func NoErrConv(s string) int64 {
	n, e := strconv.ParseInt(s, 10, 64)
	if e != nil {
		return 0
	}
	return n
}
