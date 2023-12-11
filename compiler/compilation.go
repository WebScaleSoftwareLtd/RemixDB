// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package compiler

import (
	"errors"
	"reflect"

	"remixdb.io/ast"
	"remixdb.io/engine"
)

// Do the compilation.
func (c *Compiler) doCompilation(contract *ast.ContractToken, s engine.Session) (reflect.Value, error) {
	// TODO
	return reflect.Value{}, errors.New("TODO")
}
