// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unsafe"

	"github.com/julienschmidt/httprouter"
	"github.com/valyala/fasthttp"
	"remixdb.io/errhandler"
)

type panicError struct {
	v any
}

func (p panicError) Error() string {
	return fmt.Sprint(p.v)
}

func panicWrap[T any](fn func(RequestCtx) (T, error)) func(RequestCtx) (T, error) {
	return func(ctx RequestCtx) (val T, err error) {
		defer func() {
			if r := recover(); r != nil {
				var ok bool
				err, ok = r.(error)
				if !ok {
					// This is a non-error type panic, so wrap it.
					err = panicError{r}
				}
			}
		}()
		return fn(ctx)
	}
}

type httprouterWrapper struct {
	w http.ResponseWriter
	r *http.Request
	p httprouter.Params
}

func (w httprouterWrapper) GetRequestHeader(name string) []byte {
	s := w.r.Header.Get(name)
	sLen := len(s)
	if sLen == 0 {
		return nil
	}
	bPtr := unsafe.StringData(s)
	return unsafe.Slice(bPtr, sLen)
}

func (w httprouterWrapper) GetRequestBody() []byte {
	defer w.r.Body.Close()
	b, err := io.ReadAll(w.r.Body)
	if err != nil {
		return []byte{}
	}
	return b
}

func (w httprouterWrapper) GetURLParam(name string) string {
	return w.p.ByName(name)
}

func (w httprouterWrapper) SetResponseHeader(name string, value []byte) {
	valS := ""
	if len(value) != 0 {
		valS = unsafe.String(&value[0], len(value))
	}
	w.w.Header().Set(name, valS)
}

func (w httprouterWrapper) SetResponseBody(statusCode int, value []byte) {
	w.w.WriteHeader(statusCode)
	w.w.Write(value)
}

func sendNetHttpJson(w http.ResponseWriter, statusCode int, v any, errHandler errhandler.Handler) {
	// Marshal the JSON.
	b, err := json.Marshal(v)
	if err != nil {
		// Log this error and recall it with a internal server error.
		errHandler.HandleError(err)
		sendNetHttpJson(w, http.StatusInternalServerError, APIError{
			Code:    "internal_server_error",
			Message: "Internal Server Error",
		}, errHandler)
		return
	}

	// Set the response headers.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Set the response body.
	_, _ = w.Write(b)
}

// Builds a httprouter route with a handler wrapper.
func buildHttpRouterRoute[T any](hn func(RequestCtx) (T, error), errHandler errhandler.Handler) httprouter.Handle {
	hn = panicWrap(hn)
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Build the wrapper.
		wrapper := httprouterWrapper{
			w: w,
			r: r,
			p: ps,
		}

		// Call the handler.
		v, err := hn(wrapper)
		if err != nil {
			// Process the error type.
			switch err.(type) {
			case APIError, *APIError:
				// Get the error.
				err2, ok := err.(APIError)
				if !ok {
					err2 = *err.(*APIError)
				}

				// Write X-RemixDB-Permissions if set.
				if err2.Permissions != nil {
					w.Header().Set("X-RemixDB-Permissions", strings.Join(err2.Permissions, ","))
				}

				// Send the error.
				sendNetHttpJson(w, err2.StatusCode, err2, errHandler)
			default:
				// Capture the error.
				errHandler.HandleError(err)

				// Process the error.
				sendNetHttpJson(w, http.StatusInternalServerError, APIError{
					Code:    "internal_server_error",
					Message: "Internal Server Error",
				}, errHandler)
			}

			// In all cases, return.
			return
		}

		// Send the response.
		sendNetHttpJson(w, http.StatusOK, v, errHandler)
	}
}

type fasthttpWrapper struct {
	*fasthttp.RequestCtx
}

func (w fasthttpWrapper) GetRequestHeader(name string) []byte {
	b := w.Request.Header.Peek(name)
	if len(b) == 0 {
		return nil
	}
	return b
}

func (w fasthttpWrapper) GetRequestBody() []byte {
	return w.Request.Body()
}

func (w fasthttpWrapper) GetURLParam(name string) string {
	return w.UserValue(name).(string)
}

func (w fasthttpWrapper) SetResponseHeader(name string, value []byte) {
	w.Response.Header.SetBytesV(name, value)
}

func (w fasthttpWrapper) SetResponseBody(statusCode int, value []byte) {
	w.Response.SetStatusCode(statusCode)
	w.Response.SetBody(value)
}

func sendFasthttpJson(ctx *fasthttp.RequestCtx, statusCode int, v any, errHandler errhandler.Handler) {
	// Marshal the JSON.
	b, err := json.Marshal(v)
	if err != nil {
		// Log this error and recall it with a internal server error.
		errHandler.HandleError(err)
		sendFasthttpJson(ctx, http.StatusInternalServerError, APIError{
			Code:    "internal_server_error",
			Message: "Internal Server Error",
		}, errHandler)
		return
	}

	// Set the response headers.
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.SetStatusCode(statusCode)

	// Set the response body.
	ctx.Response.SetBody(b)
}

// Builds a fasthttp route with a handler wrapper.
func buildFasthttpRoute[T any](hn func(RequestCtx) (T, error), errHandler errhandler.Handler) fasthttp.RequestHandler {
	hn = panicWrap(hn)
	return func(ctx *fasthttp.RequestCtx) {
		// Wrap the request context and call the handler.
		v, err := hn(fasthttpWrapper{ctx})
		if err != nil {
			// Process the error type.
			switch err.(type) {
			case APIError, *APIError:
				// Get the error.
				err2, ok := err.(APIError)
				if !ok {
					err2 = *err.(*APIError)
				}

				// Write X-RemixDB-Permissions if set.
				if err2.Permissions != nil {
					ctx.Response.Header.Set("X-RemixDB-Permissions", strings.Join(err2.Permissions, ","))
				}

				// Send the error.
				sendFasthttpJson(ctx, err2.StatusCode, err2, errHandler)
			default:
				// Capture the error.
				errHandler.HandleError(err)

				// Process the error.
				sendFasthttpJson(ctx, http.StatusInternalServerError, APIError{
					Code:    "internal_server_error",
					Message: "Internal Server Error",
				}, errHandler)
			}

			// In all cases, return.
			return
		}

		// Send the response.
		sendFasthttpJson(ctx, http.StatusOK, v, errHandler)
	}
}
