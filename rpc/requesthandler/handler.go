// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package requesthandler

import (
	"reflect"

	"remixdb.io/compiler"
	"remixdb.io/engine"
	"remixdb.io/rpc"
)

// Handler is used to define the request handler for the RPC.
type Handler struct {
	// Engine is used to define the engine which is powering this request.
	Engine engine.Engine

	// Compiler is used to define the compiler which is used to compile contracts.
	Compiler *compiler.Compiler
}

// Handle is used to define the request handler.
func (h Handler) Handle(partition string) (rpc.PartitionHandler, error) {
	// Create the session so we can check the partition and then use it later.
	s, err := h.Engine.CreateSession(partition)
	if err != nil {
		if err == engine.ErrPartitionDoesNotExist {
			// These should both be nil.
			return nil, nil
		}

		return nil, err
	}

	// Return the handler.
	hn := partitionHn{Engine: h.Engine, s: s, c: h.Compiler, p: partition}
	return hn.do, nil
}

type partitionHn struct {
	engine.Engine

	s engine.Session
	c *compiler.Compiler
	p string
}

func (e partitionHn) do(ctx *rpc.RequestCtx) (*rpc.Response, error) {
	// Get the API key from the map.
	apiKey, ok := ctx.AuthData["api_key"]
	if !ok {
		_ = e.s.Close()
		return rpc.RemixDBException(
			400, "missing_api_key", "The API key is missing from the request."), nil
	}

	// Handle checking the API key.
	_, permissions, err := e.GetAuthenticationPermissionsByAPIKey(ctx.Partition, apiKey)
	if err != nil {
		_ = e.s.Close()
		return nil, err
	}
	if permissions == nil {
		_ = e.s.Close()
		return rpc.RemixDBException(400, "invalid_api_key", "The API key is invalid."), nil
	}

	// Get the contract.
	contract, err := e.s.GetContractByKey(ctx.Method)
	if err != nil {
		_ = e.s.Close()
		if err == engine.ErrNotExists {
			return rpc.RemixDBException(
				404, "contract_does_not_exist", "The contract does not exist.",
			), nil
		}
		return nil, err
	}

	// Call the compiler.
	reflectValue, err := e.c.Compile(contract, e.s, e.p)
	if err != nil {
		_ = e.s.Close()
		return nil, err
	}

	// Call the contract.
	pluginRpcStructure := &pluginFriendlyRpc{
		Session: e.s,
		req:     ctx,
		perms:   permissions,
	}
	resValues := reflectValue.Call([]reflect.Value{reflect.ValueOf(pluginRpcStructure)})
	err, _ = resValues[0].Interface().(error)
	if err != nil {
		_ = e.s.Close()
		return nil, err
	}

	// Return the response.
	return pluginRpcStructure.resp, nil
}
