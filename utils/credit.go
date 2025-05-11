package utils

func CheckCredit(cvv int32, cardNumber string, year int32, month int32) bool {
	// 基本格式验证
	if !validateBasicFormat(cvv, cardNumber, year, month) {
		return false
	}

	// Luhn算法验证
	if !ValidateCreditCardWithLuhn(cardNumber) {
		return false
	}

	// 过期验证
	if !checkExpirationDate(year, month) {
		return false
	}
	return true
}
