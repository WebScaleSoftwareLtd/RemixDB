// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package localfs

import (
	"os"
	"path/filepath"

	"remixdb.io/engine/localfs/acid"
)

func integrityCheck(path string) {
	// Get the transactions folder.
	files, err := os.ReadDir(filepath.Join(path, "transactions"))
	if err != nil {
		if os.IsNotExist(err) {
			// No transactions folder, so nothing to do.
			return
		}
		panic(err)
	}

	// Go through each transaction.
	for _, dir := range files {
		if dir.IsDir() {
			// Attempt a recovery.
			tx := acid.RecoverFailedTransaction(path, dir.Name())
			if tx != nil {
				// Commit this transaction.
				if err := tx.Commit(); err != nil {
					panic(err)
				}
			}
		}
	}
}
