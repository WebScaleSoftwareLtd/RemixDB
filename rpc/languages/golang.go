// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package languages

import (
	_ "embed"

	"remixdb.io/rpc/structure"
)

//go:embed preamble/golang.go.tmpl
var golangTemplate string

func compileGo(base *structure.Base) (string, error) {
	// TODO
}

func golang(base *structure.Base) (map[Extension]string, error) {
	s, err := compileGo(base)
	if err != nil {
		return nil, err
	}
	return map[Extension]string{"go": s}, nil
}

var _ = initLanguage("golang", golang)
