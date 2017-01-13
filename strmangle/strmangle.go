package strmangle

import (
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// RandString generates a random string of length len using the chars in letterRunes
func RandString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// EnvAppName converts the app name to an environment variable compatible name, eg:
// "My-app Name" -> "MY_APP_NAME"
func EnvAppName(s string) string {
	return strings.ToUpper(replaceNonAlpha(s, '_'))
}

// DBAppName converts the app name to a database compatible name, eg:
// "My-app Name" -> "my_app_name"
func DBAppName(s string) string {
	return strings.ToLower(replaceNonAlpha(s, '_'))
}

// SanitizeAppName converts the app name to a github sanitized name, eg:
// "My_app@Name" -> "My-app-Name"
func SanitizeAppName(s string) string {
	return replaceNonAlpha(s, '-')
}

// replaceNonAlpha replaces non alphabet characters with the replace byte.
func replaceNonAlpha(s string, replace byte) string {
	byts := []byte(s)
	newByts := make([]byte, len(byts))

	for i := 0; i < len(byts); i++ {
		if byts[i] < 'A' || (byts[i] > 'Z' && byts[i] < 'a') || byts[i] > 'z' {
			newByts[i] = replace
		} else {
			newByts[i] = byts[i]
		}
	}

	return string(newByts)
}
