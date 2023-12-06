// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package singleinstance

import (
	"os"
	"path/filepath"

	"github.com/juju/fslock"
)

// EnsureSingleInstance ensures that only one instance of the program is running.
// If it isn't, it will close the program.
func EnsureSingleInstance() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	remixdbDir := filepath.Join(homedir, ".remixdb")
	if err = os.MkdirAll(remixdbDir, 0755); err != nil {
		panic(err)
	}

	lockFile := filepath.Join(remixdbDir, "lock")
	err = fslock.New(lockFile).TryLock()
	if err == nil {
		return
	}

	_, _ = os.Stderr.WriteString("Another instance of remixdb is already running.\n")
	os.Exit(1)
}
