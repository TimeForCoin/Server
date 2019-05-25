package libs

import (
	"regexp"
)

func IsEmail(email string) bool {
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

func IsGender(gender string) bool {
	Genders := []string{"man", "woman", "other"}
	for _, g := range Genders {
		if gender == g {
			return true
		}
	}
	return false
}