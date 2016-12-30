package cmd

import (
	"strings"
	"text/template"
)

var templateFuncs = template.FuncMap{
	"titleCase": func(s string) string {
		return strings.ToUpper(s)
	},
}
