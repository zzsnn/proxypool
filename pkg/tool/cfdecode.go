package tool

import (
	"bytes"
	"strconv"
	"strings"
)

func GetCFPayload(str string) string {
	s := strings.Split(str, "data-cfemail=")
	if len(s) > 1 {
		s = strings.Split(s[1], "\"")
		str = s[1]
		return str
	}
	return ""
}

func CFDecode(a string) (s string) {
	if a == "" {
		return
	}
	var e bytes.Buffer
	r, _ := strconv.ParseInt(a[0:2], 16, 0)
	for n := 4; n < len(a)+2; n += 2 {
		i, _ := strconv.ParseInt(a[n-2:n], 16, 0)
		e.WriteString(string(i ^ r))
	}
	return e.String()
}
