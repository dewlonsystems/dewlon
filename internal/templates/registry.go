package templates

import "fmt"

type Template struct {
	Name string
	URL  string
}

var Registry = map[string]Template{
	"fiber": {
		Name: "Fiber API",
		URL:  "https://github.com/dewlonsystems/fiber-template/archive/refs/heads/main.zip",
	},
}

func Get(name string) (Template, error) {

	template, ok := Registry[name]
	if !ok {
		return Template{}, fmt.Errorf("template '%s' not found", name)
	}

	return template, nil
}
