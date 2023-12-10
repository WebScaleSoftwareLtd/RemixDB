// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package webserver

import (
	"io/fs"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"
	"github.com/julienschmidt/httprouter"
	"remixdb.io/frontend"
)

func mapFrontendStaticFiles(r *httprouter.Router, dist fs.FS) {
	// Scope to the dist directory.
	var err error
	dist, err = fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}

	// Map the static files to routes.
	var processDir func(cursor fs.FS, routePrefix string)
	processDir = func(cursor fs.FS, routePrefix string) {
		dir, err := fs.ReadDir(cursor, "."+routePrefix)
		if err != nil {
			panic(err)
		}

		for _, file := range dir {
			// If route prefix is /, ignore both index.html and MAKE_GO_NOT_ERROR.
			name := file.Name()
			if routePrefix == "/" && (name == "index.html" || name == "MAKE_GO_NOT_ERROR") {
				// We do not want to handle these files here.
				continue
			}

			// Handle if this is a directory.
			if file.IsDir() {
				// Recurse into the directory.
				processDir(cursor, routePrefix+name+"/")
				continue
			}

			// Read the file from the filesystem.
			b, err := fs.ReadFile(cursor, routePrefix+name)
			if err != nil {
				panic(err)
			}

			// Get the mime type.
			mimeObj := mimetype.Detect(b)
			mimeS := mimeObj.String()
			if mimeS == "application/octet-stream" {
				// Try and guess the mime type from the file extension.
				mimeS = mime.TypeByExtension(filepath.Ext(name))
				if mimeS == "" {
					// Default to application/octet-stream.
					mimeS = "application/octet-stream"
				}
			}

			// Map the file to the route.
			r.GET(routePrefix+name, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
				w.Header().Set("Content-Type", mimeS)
				w.Header().Set("Cache-Control", "public, max-age=31536000")
				_, _ = w.Write(b)
			})
		}
	}
	processDir(dist, "/")
}

func mapFrontendIndexHtml(r *httprouter.Router, route string, indexHtml []byte) {
	r.GET(route, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(indexHtml)
	})
}

func (w *WebServer) generateHttpRoutes(containsRpcRoutes bool) http.Handler {
	r := httprouter.New()

	if containsRpcRoutes {
		r.POST("/rpc/:method", w.rpcServer.NetHTTPHandler)
		r.GET("/rpc", w.rpcServer.NetHTTPHandler)
	}

	frontendHost := os.Getenv("REMIXDB_DEV_FRONTEND_HOST")
	if frontendHost == "" {
		// Load in the index.html file.
		indexHtml, err := frontend.Dist.ReadFile("dist/index.html")
		if err != nil {
			indexHtml = []byte("RemixDB frontend not compiled into this binary!")
		}

		// Map the static files to routes.
		mapFrontendStaticFiles(r, frontend.Dist)

		// Serve the index.html file on the frontend routes.
		for _, route := range frontend.Routes {
			mapFrontendIndexHtml(r, route, indexHtml)
		}
	} else {
		// Reverse proxy the frontend to the REMIXDB_DEV_FRONTEND_HOST if there is no route.
		uri, err := url.Parse(frontendHost)
		if err != nil {
			// Set the URL scheme to http.
			uri, err = url.Parse("http://" + frontendHost)
		}
		if err != nil {
			panic(err)
		}
		r.NotFound = httputil.NewSingleHostReverseProxy(uri)
	}

	return r
}
