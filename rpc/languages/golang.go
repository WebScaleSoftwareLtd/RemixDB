// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package languages

import (
	_ "embed"

	"remixdb.io/rpc/structure"
)

//go:embed templates/golang.tmpl
var golangTemplate string

func golang(base *structure.Base, opts map[string]string) (map[Extension]string, error) {
	s, err := processGoTemplate(base, "golang", golangTemplate, nil, opts)
	if err != nil {
		return nil, err
	}
	return map[Extension]string{"go": s}, nil
}

var _ = initLanguage("golang", golang, map[string]Option{
	"package": {
		Optional: false,
		Default:  ptr("rpc"),
	},
})
