package tools

import "strconv"

// Converts string to int64. Returns 0 if error
func NoErrConv(s string) int64 {
	n, e := strconv.ParseInt(s, 10, 64)
	if e != nil {
		return 0
	}
	return n
}
