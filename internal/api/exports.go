// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package api

import "remixdb.io/internal/errhandler"

// ServerInfoV1 is the server info.
type ServerInfoV1 struct {
	// Version is the server version.
	Version string `json:"version"`

	// Hostname is the hostname of the server.
	Hostname string `json:"hostname"`

	// HostID is the host ID.
	HostID string `json:"host_id"`

	// Uptime is the server uptime in unix seconds.
	Uptime int64 `json:"uptime"`
}

// User is a user object.
type User struct {
	SudoPartition bool     `json:"sudo_partition"`
	Username      string   `json:"username"`
	Permissions   []string `json:"permissions"`
}

// MetricsV1 is the metrics.
type MetricsV1 struct {
	RAMMegabytes uint64 `json:"ram_mb"`
	Goroutines   int    `json:"goroutines"`
	GCS          int    `json:"gcs"`
}

// CreatePartitionV1Body is the body for the CreatePartitionV1 endpoint.
type CreatePartitionV1Body struct {
	SudoAPIKey    string `json:"sudo_api_key"`
	Username      string `json:"username"`
	SudoPartition bool   `json:"sudo_partition"`
}

// APIImplementation is the interface for an API implementation.
type APIImplementation interface {
	// GetServerInfoV1 returns the server info.
	GetServerInfoV1(ctx RequestCtx) (ServerInfoV1, error)

	// GetSelfUserV1 returns the self user.
	GetSelfUserV1(ctx RequestCtx) (User, error)

	// GetMetricsV1 returns the metrics.
	GetMetricsV1(ctx RequestCtx) (MetricsV1, error)

	// GetPartitionCreatedStateV1 returns the partition created state. This endpoint
	// does not require authentication.
	GetPartitionCreatedStateV1(ctx RequestCtx) (bool, error)

	// CreatePartitionV1 creates up the partition. Returns a API error with the code
	// 'partition_already_exists' if the partition already exists. The string returned
	// is the API key for the newly created partition user.
	//
	// Expected body type (JSON): CreatePartitionV1Body
	CreatePartitionV1(ctx RequestCtx) (string, error)
}

// RequestCtx is the context for a request.
type RequestCtx interface {
	// GetRequestHeader returns the value of the specified request header. The response
	// must not be mutated and only lives for the duration of the request.
	GetRequestHeader(name string) []byte

	// GetRequestBody returns the request body. The response must not be mutated and
	// only lives for the duration of the request.
	GetRequestBody() []byte

	// GetURLParam returns the value of the specified URL parameter.
	GetURLParam(name string) string

	// SetResponseHeader sets the value of the specified response header. The value
	// must not be mutated after this call.
	SetResponseHeader(name string, value []byte)

	// SetResponseBody sets the response body. The value must not be mutated after
	// this call.
	SetResponseBody(statusCode int, value []byte)
}

// APIError is used to represent an API error.
type APIError struct {
	// StatusCode is the HTTP status code.
	StatusCode int `json:"-"`

	// Permissions is set when X-RemixDB-Permissions should be set.
	Permissions []string `json:"-"`

	// Code is the error code.
	Code string `json:"code"`

	// Message is the error message.
	Message string `json:"message"`
}

// Error returns the error message.
func (e APIError) Error() string {
	return e.Message
}

// Server is the structure used to implement the API.
type Server struct {
	impl       APIImplementation
	errHandler errhandler.Handler
}

// NewServer returns a new Server.
func NewServer(impl APIImplementation, errHandler errhandler.Handler) Server {
	return Server{
		impl:       impl,
		errHandler: errHandler,
	}
}
