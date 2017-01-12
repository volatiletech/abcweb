package cmd

import (
	"math/rand"
	"strings"
	"text/template"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var templateFuncs = template.FuncMap{
	"randString": randString,
	"envAppName": envAppName,
	"dbAppName":  dbAppName,
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// envAppName converts the app name to an environment variable compatible name, eg:
// "My-app Name" -> "MY_APP_NAME"
func envAppName(s string) string {
	return strings.ToUpper(replaceNonAlpha(s, '_'))
}

// dbAppName converts the app name to a database compatible name, eg:
// "My-app Name" -> "my_app_name"
func dbAppName(s string) string {
	return strings.ToLower(replaceNonAlpha(s, '_'))
}

// sanitizeAppName converts the app name to a github sanitized name, eg:
// "My_app@Name" -> "My-app-Name"
func sanitizeAppName(s string) string {
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
