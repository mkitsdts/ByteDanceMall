package utils

import "strings"

func RewriteSentence(input string) string {
	// 跳过句子中的空白部分以及换行符
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, "\n", " ")
	return input
}
