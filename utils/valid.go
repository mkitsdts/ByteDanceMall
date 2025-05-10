package utils

import "strings"

func IsValidEmail(email string) bool {
	// 简单的邮箱格式验证
	if len(email) < 5 || len(email) > 50 {
		return false
	}
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return false
	}
	return true
}
