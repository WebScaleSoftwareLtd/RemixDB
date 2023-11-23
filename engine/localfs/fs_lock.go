// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package localfs

import (
	"os"

	"github.com/juju/fslock"
)

const lockErr = "RemixDB is already running. Please close the other instance and try again.\n"

// Acquires the filesystem lock. This is important to prevent multiple instances of RemixDB from
// running at the same time.
func mustAcquireFilesystemLock(fp string) {
	if err := fslock.New(fp).TryLock(); err != nil {
		// Exit the application with a status code 1.
		_, _ = os.Stderr.WriteString(lockErr)
		os.Exit(1)
	}
}
