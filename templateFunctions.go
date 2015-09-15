package heyfyi

import (
	"html/template"
	"strconv"
	"strings"
)

var funcMap = template.FuncMap{
	"GetViewFactUrl":             GetViewFactUrl,
	"GetSignUpUrl":               GetSignUpUrl,
	"GetSignInUrl":               GetSignInUrl,
	"GetSignOutUrl":              GetSignOutUrl,
	"GetCreateFactUrl":           GetCreateFactUrl,
	"GetHomeUrl":                 GetHomeUrl,
	"GetListFactUrl":             GetListFactUrl,
	"GetRequestPasswordResetUrl": GetRequestPasswordResetUrl,
	"GetDeleteFactUrl":           GetDeleteFactUrl,

	"TruncateString": TruncateString,
} //this provides templates with the ability to run useful functions

func GetViewFactUrl(factId int64) string {
	if factId == 0 {
		return "#error"
	}
	return ViewFactUrl.Make("factId", strconv.FormatInt(factId, 10))
}

func GetSignUpUrl() string {
	return SignUpUrl.Make()
}

func GetSignInUrl() string {
	return SignInUrl.Make()
}

func GetSignOutUrl() string {
	return SignOutUrl.Make()
}

func GetCreateFactUrl() string {
	return CreateFactUrl.Make()
}

func GetHomeUrl() string {
	return HomeUrl.Make()
}

func GetListFactUrl() string {
	return ListFactUrl.Make()
}

func GetRequestPasswordResetUrl() string {
	return RequestPasswordResetUrl.Make()
}

func GetDeleteFactUrl(factId int64) string {
	return DeleteFactUrl.Make("factId", strconv.FormatInt(factId, 10))
}

//Smart truncation function
func TruncateString(s string, charLimit int) string {
	if len(s) < charLimit {
		return s
	}
	//Return up to the last complete word + "..."
	s2 := s[:charLimit-3]
	s2ind := strings.LastIndex(s2, " ")
	if s2ind >= 0 {
		return s2[:s2ind] + "..."
	}
	//couldn't find a word boundary
	return s2 + "..."

}
