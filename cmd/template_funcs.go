package cmd

import (
	"text/template"

	"github.com/volatiletech/abcweb/strmangle"
)

var templateFuncs = template.FuncMap{
	"randString": strmangle.RandString,
	"envAppName": strmangle.EnvAppName,
	"dbAppName":  strmangle.DBAppName,
}
