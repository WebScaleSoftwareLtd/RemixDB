// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package webserver

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (w *WebServer) generateHttpRoutes(containsRpcRoutes bool) http.Handler {
	r := httprouter.New()

	if containsRpcRoutes {
		rpcHandler := func(wr http.ResponseWriter, r *http.Request, p httprouter.Params) {
			w.rpcServerLock.RLock()
			s := w.rpcServer
			w.rpcServerLock.RUnlock()

			s.NetHTTPHandler(wr, r, p)
		}
		r.POST("/rpc/:method", rpcHandler)
		r.GET("/rpc", rpcHandler)
	}

	return r
}
