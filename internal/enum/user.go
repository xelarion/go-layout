package enum

import "github.com/xelarion/go-layout/pkg/errs"

type UserGender int8

const (
	UserGenderUnknown UserGender = iota // 未知
	UserGenderMale                      // 男
	UserGenderFemale                    // 女
)

var UserGenderSlug = map[UserGender]string{
	UserGenderUnknown: "unknown",
	UserGenderMale:    "male",
	UserGenderFemale:  "female",
}

var UserGenderText = map[UserGender]string{
	UserGenderUnknown: "未知",
	UserGenderMale:    "男",
	UserGenderFemale:  "女",
}

func GetUserGender(slug string) (UserGender, error) {
	for k, v := range UserGenderSlug {
		if v == slug {
			return k, nil
		}
	}
	return 0, errs.NewBusiness("invalid UserGender")
}

func (e UserGender) Slug() string {
	if s, ok := UserGenderSlug[e]; ok {
		return s
	}
	return ""
}

func (e UserGender) Text() string {
	if s, ok := UserGenderText[e]; ok {
		return s
	}
	return ""
}
