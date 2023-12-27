// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package frontend

import (
	"embed"
	"encoding/json"
	"sort"
)

// Dist is the frontend distribution.
//
//go:embed dist
var Dist embed.FS

//go:embed routes.json
var routesBytes []byte

// Routes is an array of routes that index.html should be served for.
var Routes []string

func init() {
	var m map[string]any
	err := json.Unmarshal(routesBytes, &m)
	if err != nil {
		panic(err)
	}
	Routes = make([]string, len(m))
	i := 0
	for k := range m {
		Routes[i] = k
		i++
	}
	sort.Strings(Routes)
}
