// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package rpc

import (
	"errors"
	"sort"

	"remixdb.io/internal/rpc/languages"
	"remixdb.io/internal/rpc/structure"
)

// Languages is all the supported language keys.
func Languages() []string {
	x := make([]string, 0, len(languages.Languages))
	for k := range languages.Languages {
		x = append(x, k)
	}
	sort.Strings(x)
	return x
}

// GetOptions is used to get the compiler options for the RPC. Returns a nil map if the language
// is not supported.
func GetOptions(language string) map[string]languages.Option {
	if x, ok := languages.Languages[language]; ok {
		return x.Options
	}
	return nil
}

// Compile is used to compile the RPC. Returns a error if the language is not supported.
func Compile(
	language string, base *structure.Base, opts map[string]string,
) (map[languages.Extension]string, error) {
	if x, ok := languages.Languages[language]; ok {
		// Make a new options.
		newOpts := map[string]string{}
		for k, v := range x.Options {
			val, ok := opts[k]
			if !ok {
				if v.Default != nil {
					val = *v.Default
					ok = true
				} else if !v.Optional {
					return nil, errors.New("missing required option: " + k)
				}
			}

			if ok {
				newOpts[k] = val
			}
		}

		// Return the compiler.
		return x.Compiler(base, newOpts)
	}
	return nil, errors.New("language not supported")
}
