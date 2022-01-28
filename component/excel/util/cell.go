package util

const LineNum = 26

// ToLine 将行转为[A-Z]格式(i从1开始)
func ToLine(i int) string {
	if i <= 0 {
		return ""
	}
	char := ""
	for i > 0 {
		var m = i % LineNum
		if m == 0 {
			m = LineNum
		}
		char = toCharStr(m) + char
		i = (i - m) / LineNum
	}
	return char
}

func toCharStr(i int) string {
	if i < 0 || i > LineNum {
		return ""
	}
	if i == 0 {
		return string(rune('A' + i))
	}
	return string(rune('A' - 1 + i))
}
