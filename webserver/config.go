// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package webserver

// HTTPSOptions is used to define HTTPS options.
type HTTPSOptions struct {
	// CertFile is used to define the path to the certificate file.
	CertFile string

	// KeyFile is used to define the path to the key file.
	KeyFile string
}

// Config is used to define a web server config.
type Config struct {
	// HTTPSOptions is used to define HTTPS options. If this isn't nil, HTTPS
	// (and since we can use HTTP/2, net/http) will be used.
	HTTPSOptions *HTTPSOptions

	// H2C is used to define if H2C should be used. If HTTPOptions is not nil,
	// this is ignored.
	H2C bool

	// Host is used to define the host to listen on. Defaults to :3469.
	Host string
}
