// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package requesthandler

import (
	"remixdb.io/engine"
	"remixdb.io/rpc"
)

// Handler is used to define the request handler for the RPC.
type Handler struct {
	// Engine is used to define the engine which is powering this request.
	Engine engine.Engine
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
	hn := partitionHn{Engine: h.Engine, s: s}
	return hn.do, nil
}

type partitionHn struct {
	engine.Engine

	s engine.Session
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

	// TODO: Call the compiler!
	return nil, nil
}
