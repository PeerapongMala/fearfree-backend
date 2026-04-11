package shared

import (
	"strings"
	"unicode"
)

// ValidatePasswordComplexity checks that the password contains at least 1 digit,
// 1 uppercase letter, and 1 special character.
func ValidatePasswordComplexity(password string) bool {
	var hasDigit, hasUpper, hasSpecial bool
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	for _, ch := range password {
		switch {
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsUpper(ch):
			hasUpper = true
		case strings.ContainsRune(specialChars, ch):
			hasSpecial = true
		}
	}
	return hasDigit && hasUpper && hasSpecial
}
