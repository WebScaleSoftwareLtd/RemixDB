// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package rpc

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"nhooyr.io/websocket"
)

type netHttpHandler struct {
	req    *http.Request
	resp   http.ResponseWriter
	method string
	sent   bool
}

func (h *netHttpHandler) Method() string {
	return h.method
}

func (h *netHttpHandler) SchemaHash() string {
	return h.req.Header.Get("X-RemixDB-Schema-Hash")
}

func (h *netHttpHandler) Body() io.ReadCloser {
	return h.req.Body
}

func (h *netHttpHandler) ReturnCustomException(code int, exceptionName string, body any) error {
	if h.sent {
		return nil
	}
	h.sent = true
	he := h.resp.Header()
	he.Set("Content-Type", "application/json")
	he.Set("X-RemixDB-Exception", exceptionName)
	h.resp.WriteHeader(code)
	return json.NewEncoder(h.resp).Encode(body)
}

func (h *netHttpHandler) ReturnRemixDBException(httpCode int, code, message string) error {
	if h.sent {
		return nil
	}
	h.sent = true
	he := h.resp.Header()
	he.Set("Content-Type", "application/json")
	h.resp.WriteHeader(httpCode)
	return json.NewEncoder(h.resp).Encode(map[string]string{
		"code":    code,
		"message": message,
	})
}

func (h *netHttpHandler) ReturnRemixBytes(code int, data []byte) {
	if h.sent {
		return
	}
	h.sent = true
	h.resp.Header().Set("Content-Type", "application/remixdb-rpc")
	h.resp.WriteHeader(code)
	_, _ = h.resp.Write(data)
}

var _ httpRequest = &netHttpHandler{}

type nhooyrWebSocketCompat struct {
	*websocket.Conn

	req *http.Request
}

func (w nhooyrWebSocketCompat) Close() error { return w.Conn.CloseNow() }

func (w nhooyrWebSocketCompat) ReadMessage() (messageType int, p []byte, err error) {
	t, p, err := w.Conn.Read(w.req.Context())
	return int(t), p, err
}

func (w nhooyrWebSocketCompat) WriteMessage(messageType int, data []byte) error {
	return w.Conn.Write(w.req.Context(), websocket.MessageType(messageType), data)
}

var _ websocketConn = nhooyrWebSocketCompat{}

// NetHTTPHandler is used to handle a HTTP request via the httprouter package.
func (s *Server) NetHTTPHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Add X-Is-RemixDB: true to the response.
	w.Header().Set("X-Is-RemixDB", "true")

	// Handle /rpc/:method requests.
	m := ps.ByName("method")
	if m != "" {
		// Check if this is not a POST request.
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte("Method not allowed"))
			return
		}

		// Check if the content type is correct.
		if r.Header.Get("Content-Type") != "application/x-remixdb-rpc-mixed" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Invalid content type"))
			return
		}

		// Handle the request.
		hn := &netHttpHandler{req: r, resp: w, method: m}
		s.handleHttpRpc(hn)
		if !hn.sent {
			w.WriteHeader(http.StatusNoContent)
		}
		return
	}

	// Check if this is a GET request.
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte("Method not allowed"))
		return
	}

	// Handle if there is no websocket upgrade.
	if r.Header.Get("Upgrade") != "websocket" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid upgrade - did you mean to connect over websocket?"))
		return
	}

	// Handle the websocket upgrade.
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		// Just ignore this because it will be handled by the websocket package.
		return
	}
	s.handleWebsocketConn(nhooyrWebSocketCompat{
		Conn: conn,
		req:  r,
	})
}

var _ httprouter.Handle = (*Server)(nil).NetHTTPHandler
