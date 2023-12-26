// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package rpc

import (
	"context"
	"encoding/binary"
	"encoding/json"

	"nhooyr.io/websocket"
)

type websocketConn interface {
	Close() error
	SetReadLimit(int64)
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
}

type websocketReqImpl struct {
	conn       websocketConn
	sent       bool
	method     string
	schemaHash string
	body       []byte
	hostname   string
}

const (
	messageBinary       = int(websocket.MessageBinary)
	maxSetupMessageSize = 1024 * 1024
)

func (r *websocketReqImpl) Hostname() string {
	return r.hostname
}

func (r *websocketReqImpl) Next() bool {
	if !r.sent {
		// Send the ok response.
		r.conn.WriteMessage(messageBinary, []byte{2})
		r.sent = true
	}

	// Wait for the next message.
	messageType, msg, err := r.conn.ReadMessage()
	if err != nil {
		return false
	}
	if messageType != messageBinary {
		_ = r.conn.Close()
		return false
	}

	// Validate the message.
	if len(msg) != 1 || msg[0] != 1 {
		_ = r.conn.Close()
		return false
	}

	// Return true since the cursor is now ready.
	return true
}

func (r *websocketReqImpl) Method() string {
	return r.method
}

func (r *websocketReqImpl) SchemaHash() string {
	return r.schemaHash
}

func (r *websocketReqImpl) Body() []byte {
	return r.body
}

func (r *websocketReqImpl) ReturnCustomException(code int, exceptionName string, body any) error {
	defer r.conn.Close()

	// Encode the body.
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// Create the message.
	b := make([]byte, 1+2+len(exceptionName)+len(bodyBytes))
	b[0] = 1 // 1 = custom exception
	binary.LittleEndian.PutUint16(b[1:], uint16(len(exceptionName)))
	copy(b[3:], exceptionName)
	copy(b[3+len(exceptionName):], bodyBytes)

	// Send the message.
	return r.conn.WriteMessage(messageBinary, b)
}

func (r *websocketReqImpl) ReturnRemixDBException(httpCode int, code, message string) error {
	defer r.conn.Close()

	// Create the message.
	b := make([]byte, 1+2+len(code)+len(message))
	// the first byte is 0, which means it is a RemixDB exception.
	binary.LittleEndian.PutUint16(b[1:], uint16(len(code)))
	copy(b[3:], code)
	copy(b[3+len(code):], message)

	// Send the message.
	return r.conn.WriteMessage(messageBinary, b)
}

func (r *websocketReqImpl) ReturnRemixBytes(code int, data []byte) {
	// Create the bytes containing the message.
	b := make([]byte, 1+len(data))
	b[0] = 2 // 2 = success
	copy(b[1:], data)

	// Send the message.
	_ = r.conn.WriteMessage(messageBinary, b)
}

func (r *websocketReqImpl) ReturnEOF() {
	_ = r.conn.WriteMessage(messageBinary, []byte{3})
	_ = r.conn.Close()
}

func (r *websocketReqImpl) Context() context.Context {
	return context.Background()
}

var _ websocketRequest = (*websocketReqImpl)(nil)

func (s *Server) handleWebsocketConn(conn websocketConn, hostname string) {
	// We always close the connection at the end.
	defer conn.Close()

	// Set the read limit to 1MB for the initial setup message.
	conn.SetReadLimit(maxSetupMessageSize)

	// Read the setup message.
	messageType, msg, err := conn.ReadMessage()
	if err != nil {
		return
	}

	// Check the message type.
	if messageType != messageBinary {
		return
	}

	// If the message length is less than 4, then it is invalid.
	if len(msg) < 4 {
		return
	}

	// Get the method.
	methodLen := int(binary.LittleEndian.Uint16(msg))
	msg = msg[2:]
	if len(msg) < methodLen {
		// This would make it invalid.
		return
	}
	method := string(msg[:methodLen])
	msg = msg[methodLen:]

	// Get the schema hash.
	if len(msg) < 2 {
		// This would make it invalid.
		return
	}
	schemaHashLen := int(binary.LittleEndian.Uint16(msg))
	msg = msg[2:]
	if len(msg) < schemaHashLen {
		// This would make it invalid.
		return
	}
	schemaHash := string(msg[:schemaHashLen])
	msg = msg[schemaHashLen:]

	// Set the read limit to 1 byte.
	conn.SetReadLimit(1)

	// Pass the wrapper through to the handler.
	s.handleRpc(&websocketReqImpl{
		conn:       conn,
		method:     method,
		schemaHash: schemaHash,
		body:       msg,
		hostname:   hostname,
	})
}
