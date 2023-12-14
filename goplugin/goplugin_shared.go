// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package goplugin

// Plugin defines the interface for a Go plugin.
type Plugin interface {
	Lookup(name string) (any, error)
}
