package cmd

import (
	"text/template"

	"github.com/volatiletech/abcweb/v5/strmangle"
)

var templateFuncs = template.FuncMap{
	"randString": strmangle.RandString,
	"envAppName": strmangle.EnvAppName,
	"dbAppName":  strmangle.DBAppName,
}
