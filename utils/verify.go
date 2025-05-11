package utils

import (
	"strconv"
	"time"
)

func validateBasicFormat(cvv int32, cardNumber string, year int32, month int32) bool {
	// CVV长度验证
	if cvv < 100 || cvv > 999 {
		return false
	}

	// 卡号长度验证
	if len(cardNumber) != 16 {
		return false
	}

	// 年份范围验证
	if year < 2010 || year > 2050 {
		return false
	}

	// 月份范围验证
	if month < 1 || month > 12 {
		return false
	}

	return true
}

func ValidateCreditCardWithLuhn(cardNumber string) bool {
	// Luhn算法验证
	sum := 0
	alt := false
	for i := len(cardNumber) - 1; i >= 0; i-- {
		n, _ := strconv.Atoi(string(cardNumber[i]))
		if alt {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		alt = !alt
	}
	return sum%10 == 0
}

func checkExpirationDate(year int32, month int32) bool {
	// 获取当前时间
	currentYear := int32(time.Now().Year())
	currentMonth := int32(time.Now().Month())

	// 检查过期日期
	if year < currentYear || (year == currentYear && month < currentMonth) {
		return false
	}
	return true
}
