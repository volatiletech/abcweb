package cmd

var templateFuncs = templates.FuncMap{
	"titleCase": func(s string) string {
		strconv.UpperCase(s)
	}
}
