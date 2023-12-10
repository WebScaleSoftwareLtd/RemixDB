// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package compiler

import (
	"reflect"

	"remixdb.io/ast"
	"remixdb.io/engine"
	"remixdb.io/goplugin"
	"remixdb.io/utils"
)

// Compiler is used to compile a contract into a Go plugin or cache it.
type Compiler struct {
	compilationCache utils.TLRUCache[string, map[string]reflect.Value]

	// GoPluginCompiler is the Go plugin compiler.
	GoPluginCompiler goplugin.GoPluginCompiler
}

// Compile is used to compile a contract into a Go plugin.
func (c *Compiler) Compile(contract *ast.ContractToken, s engine.Session, partition string) (reflect.Value, error) {
	// Try and load from the cache.
	compiledItems, ok := c.compilationCache.Get(partition)
	if ok {
		// Try and load the contract from the cache.
		compiledItem, ok := compiledItems[contract.Name]
		if ok {
			return compiledItem, nil
		}

		// Make compiledItems nil to prevent the risk of the cache being mutated.
		compiledItems = nil
	}

	// TODO
	return reflect.Value{}, nil
}
