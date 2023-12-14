// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package goplugin

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cp "github.com/otiai10/copy"
	"remixdb.io/logger"
)

// SetupGoCompilerForTesting sets up a Go compiler for testing.
func SetupGoCompilerForTesting(t *testing.T) GoPluginCompiler {
	// Setup the temporary directory used for the tests.
	tempDir := t.TempDir()

	// Get the plugin path that the application uses.
	path := os.Getenv("REMIXDB_GOPLUGIN_PATH")
	if path == "" {
		// Go ahead and set the path to the user RemixDB directory.
		homedir, err := os.UserHomeDir()
		if err != nil {
			t.Fatal(err)
		}
		path = filepath.Join(homedir, ".remixdb", "goplugin")
	}

	// Check if the path exists.
	s, err := os.Stat(path)
	if err != nil {
		t.Fatal(fmt.Errorf("failed to stat %s: %w - please run the setup file first", path, err))
	}

	// Check if the path is a directory.
	if !s.IsDir() {
		t.Fatal(fmt.Errorf("%s is not a directory - please run the setup file first", path))
	}

	// Copy the files from the path into the temporary directory.
	if err = cp.Copy(path, tempDir); err != nil {
		t.Fatal(err)
	}

	// Create the compiler.
	return NewGoPluginCompiler(logger.NewTestingLogger(t), tempDir)
}
