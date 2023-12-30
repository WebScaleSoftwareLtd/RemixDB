// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package dbstart

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/common-nighthawk/go-figure"
)

func printSplashScreen() {
	version := ""
	info, ok := debug.ReadBuildInfo()
	if ok {
		rev := ""
		dirty := false
		for _, kv := range info.Settings {
			switch kv.Key {
			case "vcs.revision":
				rev = kv.Value
			case "vcs.modified":
				dirty = kv.Value == "true"
			}
		}
		if rev != "" {
			version = rev
			if dirty {
				version += " (changes since git revision)"
			}
		}
	}

	if version == "" {
		version = "unknown"
	}

	s := figure.NewFigure("RemixDB", "", true).String()
	fmt.Println(s + `

Version: ` + version + `
Platform: ` + runtime.GOOS + `
Architecture: ` + runtime.GOARCH + `
`)
}
