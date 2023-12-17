// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"remixdb.io/errhandler"
	"remixdb.io/logger"
)

func Test_panicError_Error(t *testing.T) {
	err := panicError{v: 123}
	assert.Equal(t, "123", err.Error())
}

func doPanicTests[T any](t *testing.T) {
	t.Helper()

	tests := []struct {
		name string

		throws    any
		wantedErr error
	}{
		{name: "no error"},
		{
			name:      "regular error",
			wantedErr: errors.New("123"),
		},
		{
			name:      "string",
			throws:    "123",
			wantedErr: panicError{v: "123"},
		},
		{
			name:      "int",
			throws:    123,
			wantedErr: panicError{v: 123},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Handle if throws is nil.
			if tt.throws == nil && tt.wantedErr != nil {
				tt.throws = tt.wantedErr
			}

			// Handle the panic wrap.
			_, err := panicWrap(func(RequestCtx) (_ T, _ error) {
				if tt.throws != nil {
					panic(tt.throws)
				}
				return
			})(nil)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func Test_panicWrap(t *testing.T) {
	t.Run("string", doPanicTests[string])
	t.Run("int", doPanicTests[int])
}

func Test_httprouterWrapper_GetRequestHeader(t *testing.T) {
	tests := []struct {
		name string

		value string
	}{
		{name: "empty", value: ""},
		{name: "non-empty", value: "123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httprouterWrapper{r: &http.Request{
				Header: http.Header{
					"Test": []string{tt.value},
				},
			}}
			v := []byte(tt.value)
			if tt.value == "" {
				v = nil
			}
			assert.Equal(t, v, w.GetRequestHeader("Test"))
		})
	}
}

type errReader struct {
	*strings.Reader
	err error
}

func (e errReader) Read(b []byte) (int, error) {
	if e.err != nil {
		return 0, e.err
	}
	return e.Reader.Read(b)
}

func (errReader) Close() error { return nil }

func Test_httprouterWrapper_GetRequestBody(t *testing.T) {
	tests := []struct {
		name string

		body string
		err  error
	}{
		{name: "empty", body: ""},
		{name: "non-empty", body: "123"},
		{name: "error", err: errors.New("123")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httprouterWrapper{r: &http.Request{
				Body: errReader{
					Reader: strings.NewReader(tt.body),
					err:    tt.err,
				},
			}}
			assert.Equal(t, []byte(tt.body), w.GetRequestBody())
		})
	}
}

func Test_httprouterWrapper_GetURLParam(t *testing.T) {
	tests := []struct {
		name string

		value string
	}{
		{name: "empty", value: ""},
		{name: "non-empty", value: "123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httprouterWrapper{p: httprouter.Params{
				httprouter.Param{Key: "Test", Value: tt.value},
			}}
			assert.Equal(t, tt.value, w.GetURLParam("Test"))
		})
	}
}

type fakeHeaderResponseWriter struct {
	http.ResponseWriter

	header http.Header
}

func (f *fakeHeaderResponseWriter) Header() http.Header {
	if f.header == nil {
		f.header = http.Header{}
	}
	return f.header
}

func Test_httprouterWrapper_SetResponseHeader(t *testing.T) {
	tests := []struct {
		name string

		value string
	}{
		{name: "empty", value: ""},
		{name: "non-empty", value: "123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httprouterWrapper{w: &fakeHeaderResponseWriter{}}
			w.SetResponseHeader("test", []byte(tt.value))
			assert.Equal(t, tt.value, w.w.Header().Get("test"))
		})
	}
}

type fakeBodyResponseWriter struct {
	http.ResponseWriter

	body       []byte
	statusCode int
}

func (f *fakeBodyResponseWriter) WriteHeader(statusCode int) {
	f.statusCode = statusCode
}

func (f *fakeBodyResponseWriter) Write(body []byte) (int, error) {
	f.body = body
	return len(body), nil
}

func Test_httprouterWrapper_SetResponseBody(t *testing.T) {
	tests := []struct {
		name string

		value      string
		statusCode int
	}{
		{name: "empty", value: "", statusCode: 123},
		{name: "non-empty", value: "123", statusCode: 456},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httprouterWrapper{w: &fakeBodyResponseWriter{}}
			w.SetResponseBody(tt.statusCode, []byte(tt.value))
			assert.Equal(t, tt.statusCode, w.w.(*fakeBodyResponseWriter).statusCode)
			assert.Equal(t, []byte(tt.value), w.w.(*fakeBodyResponseWriter).body)
		})
	}
}

type fakeBodyAndHeadersCombined struct {
	fakeBodyResponseWriter
	fakeHeaderResponseWriter
}

func Test_sendNetHttpJson(t *testing.T) {
	tests := []struct {
		name string

		body       any
		statusCode int

		wantedBody []byte
	}{
		{
			name:       "string",
			body:       "123",
			statusCode: 456,
			wantedBody: []byte(`"123"`),
		},
		{
			name:       "int",
			body:       123,
			statusCode: 456,
			wantedBody: []byte(`123`),
		},
		{
			name:       "nil",
			body:       nil,
			statusCode: 456,
			wantedBody: []byte(`null`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &fakeBodyAndHeadersCombined{}
			errHandler := errhandler.Handler{
				Logger: logger.NewTestingLogger(t),
			}
			sendNetHttpJson(w, tt.statusCode, tt.body, errHandler)
			assert.Equal(t, tt.statusCode, w.statusCode)
			assert.Equal(t, tt.wantedBody, w.body)
			assert.Equal(t, "application/json", w.header.Get("Content-Type"))
		})
	}
}

func Test_buildHttpRouterRoute(t *testing.T) {
	tests := []struct {
		name string

		err    error
		panics bool

		wantHeaders map[string]string
	}{
		{name: "no error"},
		{name: "error", err: errors.New("123")},
		{name: "panic", err: errors.New("123"), panics: true},

		{
			name: "api error",
			err: APIError{
				StatusCode: 456,
				Code:       "789",
				Message:    "hello world",
			},
		},
		{
			name: "api error with permissions",
			err: &APIError{
				StatusCode:  456,
				Code:        "789",
				Message:     "hello world",
				Permissions: []string{"abc", "def"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get the error handler.
			errHandler := errhandler.Handler{Logger: logger.NewTestingLogger(t)}

			// Defines the http body.
			body := []byte("testing testing")

			// Defines the http request.
			req, err := http.NewRequest(http.MethodPost, "http://example.com/test", io.NopCloser(bytes.NewReader(body)))
			if err != nil {
				require.NoError(t, err)
			}

			// Defines the http response.
			resp := &fakeBodyAndHeadersCombined{}

			// Defines the httprouter params.
			params := httprouter.Params{
				httprouter.Param{Key: "test", Value: "123"},
			}

			// Defines the handler.
			handler := buildHttpRouterRoute(func(r RequestCtx) (string, error) {
				// Check the URL param is there as expected.
				assert.Equal(t, "123", r.GetURLParam("test"))

				// Check the body is there as expected.
				assert.Equal(t, body, r.GetRequestBody())

				// Check if we are meant to error.
				if tt.err != nil {
					// Check if we should panic it.
					if tt.panics {
						panic(tt.err)
					}

					// Return the error.
					return "", tt.err
				}

				// Return the response.
				return "testing testing", nil
			}, errHandler)

			// Call the handler.
			handler(resp, req, params)

			// Check the content type.
			assert.Equal(t, "application/json", resp.header.Get("Content-Type"))

			if tt.err == nil {
				// Check the response body.
				assert.Equal(t, []byte(`"testing testing"`), resp.body)

				// Check the status code.
				assert.Equal(t, http.StatusOK, resp.statusCode)

				// Return here.
				return
			}

			// Check if the error is a API error.
			err2, ok := tt.err.(*APIError)
			if !ok {
				// Try APIError.
				v, ok := tt.err.(APIError)
				if ok {
					err2 = &v
				}
			}

			// If this is a API error, make sure the response is correct.
			if err2 != nil {
				// Check the body and status.
				b, err := json.Marshal(err2)
				require.NoError(t, err)
				assert.Equal(t, b, resp.body)
				assert.Equal(t, err2.StatusCode, resp.statusCode)

				// Check the permissions.
				if err2.Permissions != nil {
					assert.Equal(t, err2.Permissions, strings.Split(resp.header.Get("X-RemixDB-Permissions"), ","))
				}

				// Return here.
				return
			}

			// Expect the body to be a internal server error.
			mustJsonString := func(v any) string {
				b, err := json.Marshal(v)
				require.NoError(t, err)
				return string(b)
			}
			assert.Equal(t, mustJsonString(APIError{
				Code:    "internal_server_error",
				Message: "Internal Server Error",
			}), string(resp.body))
			assert.Equal(t, http.StatusInternalServerError, resp.statusCode)
		})
	}
}

func Test_fasthttpWrapper_GetRequestHeader(t *testing.T) {
	tests := []struct {
		name string

		value string
	}{
		{name: "empty", value: ""},
		{name: "non-empty", value: "123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.Header.Set("Test", tt.value)
			w := fasthttpWrapper{ctx}
			v := []byte(tt.value)
			if tt.value == "" {
				v = nil
			}
			assert.Equal(t, v, w.GetRequestHeader("Test"))
		})
	}
}

func Test_fasthttpWrapper_GetRequestBody(t *testing.T) {
	tests := []struct {
		name string

		body string
		err  error
	}{
		{name: "empty", body: ""},
		{name: "non-empty", body: "123"},
		{name: "error", err: errors.New("123")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.SetBodyStream(errReader{
				Reader: strings.NewReader(tt.body),
				err:    tt.err,
			}, -1)
			w := fasthttpWrapper{ctx}
			assert.Equal(t, []byte(tt.body), w.GetRequestBody())
		})
	}
}

func Test_fasthttpWrapper_GetURLParam(t *testing.T) {
	tests := []struct {
		name string

		value string
	}{
		{name: "empty", value: ""},
		{name: "non-empty", value: "123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			ctx.SetUserValue("Test", tt.value)
			w := fasthttpWrapper{ctx}
			assert.Equal(t, tt.value, w.GetURLParam("Test"))
		})
	}
}

func Test_fasthttpWrapper_SetResponseHeader(t *testing.T) {
	tests := []struct {
		name string

		value string
	}{
		{name: "empty", value: ""},
		{name: "non-empty", value: "123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			w := fasthttpWrapper{ctx}
			w.SetResponseHeader("test", []byte(tt.value))
			assert.Equal(t, tt.value, string(ctx.Response.Header.Peek("test")))
		})
	}
}

func Test_fasthttpWrapper_SetResponseBody(t *testing.T) {
	tests := []struct {
		name string

		value      string
		statusCode int
	}{
		{name: "empty", value: "", statusCode: 123},
		{name: "non-empty", value: "123", statusCode: 456},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			w := fasthttpWrapper{ctx}
			w.SetResponseBody(tt.statusCode, []byte(tt.value))
			assert.Equal(t, tt.statusCode, ctx.Response.StatusCode())
			assert.Equal(t, []byte(tt.value), ctx.Response.Body())
		})
	}
}

func Test_sendFasthttpJson(t *testing.T) {
	tests := []struct {
		name string

		body       any
		statusCode int

		wantedBody []byte
	}{
		{
			name:       "string",
			body:       "123",
			statusCode: 456,
			wantedBody: []byte(`"123"`),
		},
		{
			name:       "int",
			body:       123,
			statusCode: 456,
			wantedBody: []byte(`123`),
		},
		{
			name:       "nil",
			body:       nil,
			statusCode: 456,
			wantedBody: []byte(`null`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			errHandler := errhandler.Handler{
				Logger: logger.NewTestingLogger(t),
			}
			sendFasthttpJson(ctx, tt.statusCode, tt.body, errHandler)
			assert.Equal(t, tt.statusCode, ctx.Response.StatusCode())
			assert.Equal(t, tt.wantedBody, ctx.Response.Body())
			assert.Equal(t, "application/json", string(ctx.Response.Header.Peek("Content-Type")))
		})
	}
}

func Test_buildFasthttpRoute(t *testing.T) {
	tests := []struct {
		name string

		err    error
		panics bool

		wantHeaders map[string]string
	}{
		{name: "no error"},
		{name: "error", err: errors.New("123")},
		{name: "panic", err: errors.New("123"), panics: true},

		{
			name: "api error",
			err: APIError{
				StatusCode: 456,
				Code:       "789",
				Message:    "hello world",
			},
		},
		{
			name: "api error with permissions",
			err: &APIError{
				StatusCode:  456,
				Code:        "789",
				Message:     "hello world",
				Permissions: []string{"abc", "def"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get the error handler.
			errHandler := errhandler.Handler{Logger: logger.NewTestingLogger(t)}

			// Defines the http body.
			body := []byte("testing testing")

			// Defines the http request.
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.Header.SetMethod(http.MethodPost)
			ctx.Request.SetRequestURI("http://example.com/test")
			ctx.Request.SetBody(body)

			// Defines the httprouter params.
			ctx.SetUserValue("test", "123")

			// Defines the handler.
			handler := buildFasthttpRoute(func(r RequestCtx) (string, error) {
				// Check the URL param is there as expected.
				assert.Equal(t, "123", r.GetURLParam("test"))

				// Check the body is there as expected.
				assert.Equal(t, body, r.GetRequestBody())

				// Check if we are meant to error.
				if tt.err != nil {
					// Check if we should panic it.
					if tt.panics {
						panic(tt.err)
					}

					// Return the error.
					return "", tt.err
				}

				// Return the response.
				return "testing testing", nil
			}, errHandler)

			// Call the handler.
			handler(ctx)

			// Check the content type.
			assert.Equal(t, "application/json", string(ctx.Response.Header.Peek("Content-Type")))

			if tt.err == nil {
				// Check the response body.
				assert.Equal(t, []byte(`"testing testing"`), ctx.Response.Body())

				// Check the status code.
				assert.Equal(t, http.StatusOK, ctx.Response.StatusCode())

				// Return here.
				return
			}

			// Check if the error is a API error.
			err2, ok := tt.err.(*APIError)
			if !ok {
				// Try APIError.
				v, ok := tt.err.(APIError)
				if ok {
					err2 = &v
				}
			}

			// If this is a API error, make sure the response is correct.
			if err2 != nil {
				// Check the body and status.
				b, err := json.Marshal(err2)
				require.NoError(t, err)
				assert.Equal(t, b, ctx.Response.Body())
				assert.Equal(t, err2.StatusCode, ctx.Response.StatusCode())

				// Check the permissions.
				if err2.Permissions != nil {
					assert.Equal(t, err2.Permissions, strings.Split(string(ctx.Response.Header.Peek("X-RemixDB-Permissions")), ","))
				}

				// Return here.
				return
			}

			// Expect the body to be a internal server error.
			mustJsonString := func(v any) string {
				b, err := json.Marshal(v)
				require.NoError(t, err)
				return string(b)
			}
			assert.Equal(t, mustJsonString(APIError{
				Code:    "internal_server_error",
				Message: "Internal Server Error",
			}), string(ctx.Response.Body()))
			assert.Equal(t, http.StatusInternalServerError, ctx.Response.StatusCode())
		})
	}
}
