// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package compiler

import (
	"context"
	"reflect"
	"runtime"
	"sync"

	"github.com/fatih/semgroup"
	"remixdb.io/internal/ast"
	"remixdb.io/internal/engine"
	"remixdb.io/internal/goplugin"
)

// Compiler is used to compile a contract into a Go plugin or cache it. Note the job of
// the compiler is not to validate the contract. You should invoke the planner before doing
// any compilation from a user input.
type Compiler struct {
	compilationCache   map[string]map[string]reflect.Value
	compilationCacheMu sync.RWMutex

	// GoPluginCompiler is the Go plugin compiler.
	GoPluginCompiler goplugin.GoPluginCompiler
}

// FlushPartitionCache is used to flush the cache for a partition.
func (c *Compiler) FlushPartitionCache(partition string) {
	c.compilationCacheMu.Lock()
	defer c.compilationCacheMu.Unlock()

	if c.compilationCache == nil {
		return
	}

	delete(c.compilationCache, partition)
}

// FlushCompiledMethodFromCache is used to flush a compiled method from the cache.
func (c *Compiler) FlushCompiledMethodFromCache(partition, contract string) {
	c.compilationCacheMu.Lock()
	defer c.compilationCacheMu.Unlock()

	if c.compilationCache == nil {
		return
	}

	compiledItems, ok := c.compilationCache[partition]
	if !ok {
		return
	}

	delete(compiledItems, contract)
}

// Compile is used to compile a contract into a Go plugin.
func (c *Compiler) Compile(contract *ast.ContractToken, s engine.Session, partition string) (reflect.Value, error) {
	// Try and load from the cache.
	c.compilationCacheMu.RLock()
	m := c.compilationCache
	if m == nil {
		m = map[string]map[string]reflect.Value{}
	}
	compiledItems, ok := m[partition]
	if ok {
		// Try and load the contract from the cache.
		compiledItem, ok := compiledItems[contract.Name]
		if ok {
			c.compilationCacheMu.RUnlock()
			return compiledItem, nil
		}

		// Make compiledItems nil to prevent the risk of the cache being mutated.
		compiledItems = nil
	}
	c.compilationCacheMu.RUnlock()

	// Do the compilation.
	compiledItem, err := c.doCompilation(contract, s)
	if err != nil {
		return reflect.Value{}, err
	}

	// Cache the compiled item.
	c.compilationCacheMu.Lock()
	defer c.compilationCacheMu.Unlock()
	m = c.compilationCache
	if m == nil {
		m = map[string]map[string]reflect.Value{}
		c.compilationCache = m
	}
	compiledItems, ok = m[partition]
	if !ok {
		compiledItems = map[string]reflect.Value{}
		m[partition] = compiledItems
	}
	compiledItems[contract.Name] = compiledItem
	return compiledItem, nil
}

func (c *Compiler) compilePartition(s engine.Session, partition string) error {
	// Defines the contract compilations.
	contractCompilations := map[string]reflect.Value{}

	// Get the contracts.
	contracts, err := s.Contracts()
	if err != nil {
		return err
	}

	// Go through each contract and compile it.
	for _, contract := range contracts {
		// Compile the contract.
		compiledItem, err := c.doCompilation(contract, s)
		if err != nil {
			return err
		}

		// Cache the compiled item.
		contractCompilations[contract.Name] = compiledItem
	}

	// Cache the compiled items.
	c.compilationCacheMu.Lock()
	defer c.compilationCacheMu.Unlock()
	m := c.compilationCache
	if m == nil {
		m = map[string]map[string]reflect.Value{}
		c.compilationCache = m
	}
	m[partition] = contractCompilations
	return nil
}

// CompileAll is used to compile all contracts in the database.
func (c *Compiler) CompileAll(e engine.Engine) error {
	partitions := e.Partitions()
	sg := semgroup.NewGroup(context.Background(), int64(10*runtime.NumCPU()))
	for _, partition := range partitions {
		partition := partition
		sg.Go(func() error {
			// Create the session.
			s, err := e.CreateSession(partition)
			if err != nil {
				if err == engine.ErrPartitionDoesNotExist {
					// We were raced by a partition deletion!
					return nil
				}
				return err
			}

			// Defer closing the session.
			defer s.Close()

			// Do the compilation.
			return c.compilePartition(s, partition)
		})
	}
	return sg.Wait()
}
