package libs

import (
	"regexp"
	"time"
)

// IsEmail 判断是否为合法邮箱
func IsEmail(email string) bool {
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// IsGender 判断是否为合法性别
func IsGender(gender string) bool {
	Genders := []string{"man", "woman", "other"}
	for _, g := range Genders {
		if gender == g {
			return true
		}
	}
	return false
}

// IsUserType 判断是否为合法用户类型
func IsUserType(userType string) bool {
	UserType := []string{"ban", "normal", "admin", "root"}
	for _, t := range UserType {
		if userType == t {
			return true
		}
	}
	return false
}

// CheckReward 判断是否为合法酬劳
func CheckReward(reward, rewardObject string, rewardValue float32) {
	if reward == "money" || reward == "rmb" {
		Assert(rewardValue != 0, "invalid_reward_value", 400)
	} else if reward == "object" {
		Assert(rewardObject != "", "invalid_reward_object", 400)
	} else {
		Assert(false, "invalid_reward", 400)
	}
}

// CheckDateDuring 判断是否为合法日期区间
func CheckDateDuring(start, end int64) {
	nowTime := time.Now()
	startDate := time.Unix(start, 0)
	endDate := time.Unix(end, 0)
	Assert(startDate.After(nowTime), "invalid_start_date", 400)
	Assert(endDate.After(startDate), "now_allow_date", 403)
}
