// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

// Package httpproxypatch is used to patch the http.DefaultTransport to use the
// system proxy settings. Importing this package will automatically make it work,
// which is useful for both the binary and the tests since they might need to be
// ran on a system with a proxy. Is this cursed? Very cursed, but there's no good
// way to handle proxies.
package httpproxypatch

import (
	"net/http"

	"github.com/mattn/go-ieproxy"
)

func init() {
	http.DefaultTransport.(*http.Transport).Proxy = ieproxy.GetProxyFunc()
}
