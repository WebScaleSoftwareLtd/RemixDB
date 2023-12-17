// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package api

import (
	"regexp"

	"github.com/fasthttp/router"
	"github.com/julienschmidt/httprouter"
	"remixdb.io/errhandler"
)

// Defines the routes for the API.
func (s Server) mapRoutes(router any) {
	// Define the server data.
	d := serverData{
		router:     router,
		errHandler: s.errHandler,
	}

	// Define the routes.
	doMapping(d, "GET", "/api/v1/info", s.impl.GetServerInfoV1)
	doMapping(d, "GET", "/api/v1/user", s.impl.GetSelfUserV1)
}

// Defines the regex to get all the {params} from a route.
var routeParamRegex = regexp.MustCompile(`{([^}]+)}`)

// Defines the server data so we aren't passing around any a bunch.
type serverData struct {
	router     any
	errHandler errhandler.Handler
}

// Does the mapping into the router.
func doMapping[T any](d serverData, method, route string, fn func(RequestCtx) (T, error)) {
	switch router := d.router.(type) {
	case *router.Router:
		router.Handle(method, route, buildFasthttpRoute(fn, d.errHandler))
	case *httprouter.Router:
		router.Handle(
			method,
			routeParamRegex.ReplaceAllString(route, ":$1"),
			buildHttpRouterRoute(fn, d.errHandler))
	default:
		panic("unknown router type")
	}
}

// AddToFasthttpRouter adds the API to the given fasthttp router.
func (s Server) AddToFasthttpRouter(router *router.Router) {
	s.mapRoutes(router)
}

// AddToHttpRouter adds the API to the given http router.
func (s Server) AddToHttpRouter(router *httprouter.Router) {
	s.mapRoutes(router)
}
