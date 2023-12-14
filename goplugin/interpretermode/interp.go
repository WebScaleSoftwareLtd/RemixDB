// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package interpretermode

import (
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// PluginLike is used to return a struct that acts plugin like.
type PluginLike struct {
	i *interp.Interpreter
}

// Lookup is used to lookup a value from the interpreter.
func (p PluginLike) Lookup(name string) (any, error) {
	return p.i.Eval(name)
}

// Parse is used to parse the code and return a plugin like struct.
func Parse(code string) (PluginLike, error) {
	i := interp.New(interp.Options{})
	_ = i.Use(stdlib.Symbols)

	_, err := i.Eval(code)
	if err != nil {
		return PluginLike{}, err
	}

	return PluginLike{i}, nil
}
