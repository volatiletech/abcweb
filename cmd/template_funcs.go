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
	"toUpper":    strings.ToUpper,
	"randString": randString,
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
