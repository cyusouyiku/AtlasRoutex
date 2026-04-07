package utils

import "regexp"

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

func IsEmail(v string) bool {
return emailRe.MatchString(v)
}
