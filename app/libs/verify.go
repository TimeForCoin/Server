package libs

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func IsID(id string) bool {
	_, err := primitive.ObjectIDFromHex(id)
	return err == nil
}

func IsUserType(userType string) bool {
	UserType := []string{"ban", "normal", "admin", "root"}
	for _, t := range UserType {
		if userType == t {
			return true
		}
	}
	return false
}