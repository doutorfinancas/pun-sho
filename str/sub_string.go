package str

import "strings"

func SubString(str string, start int, end int) string {
	wb := strings.Split(str, "")

	if start < 0 || end < 0 {
		return ""
	}

	if len(wb) < start {
		return ""
	}

	if len(wb) < end {
		return strings.Join(wb[start:], "")
	}

	return strings.Join(wb[start:end], "")
}
