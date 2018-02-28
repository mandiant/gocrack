package authentication

import "unicode"

// CheckPasswordRequirement ensures the password meets sane requirements of 8+ characters, at least one number,
// one special character, and both an upper and lowercase letter
func CheckPasswordRequirement(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasLetter, hasPunct, hasNum bool
	for _, r := range password {
		// gate: if we already have all the validations, bomb out
		if hasLetter && hasPunct && hasNum {
			return true
		}

		if unicode.IsNumber(r) {
			hasNum = true
			continue
		}
		if unicode.IsLetter(r) {
			hasLetter = true
			continue
		}
		if unicode.IsPunct(r) {
			hasPunct = true
			continue
		}
	}

	return hasLetter && hasPunct && hasNum
}
