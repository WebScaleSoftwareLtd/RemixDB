// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package ast_test

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"remixdb.io/ast"
	"remixdb.io/utils"
)

var (
	startPrefix        = filepath.Join(".", "testdata", "tests")
	resultsReplacement = filepath.Join(".", "testdata", "results")
)

func getResultsPath(fp string) string {
	prev := fp[len(startPrefix):]

	// If it ends .rql, remove it.
	if filepath.Ext(prev) == ".rql" {
		prev = prev[:len(prev)-4]
	}

	return resultsReplacement + prev + ".txt"
}

func processFile(fp string, t *testing.T) {
	// Read the file.
	b, err := os.ReadFile(fp)
	if err != nil {
		t.Fatal(err)
	}

	// Parse the file.
	r, perr := ast.Parse(string(b))
	var toString any
	if perr == nil {
		toString = r
	} else {
		toString = perr
	}

	// Turn it into a spew string.
	current := spew.Sdump(toString)

	// If RESULTS_UPDATE is set to 1, write the current results to the results file.
	resPath := getResultsPath(fp)
	if os.Getenv("RESULTS_UPDATE") == "1" {
		// Ensure the directory exists.
		err = os.MkdirAll(filepath.Dir(resPath), 0755)
		if err != nil {
			t.Fatal(err)
		}

		// Write the results.
		err = os.WriteFile(resPath, []byte(current), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Read the results.
	b, err = os.ReadFile(resPath)
	if err != nil {
		t.Fatal(err)
	}

	// Compare the results.
	assert.Equal(t, string(b), current)
}

func processFolder(fp string, t *testing.T) {
	// Read the directory.
	dir, err := os.ReadDir(fp)
	if err != nil {
		t.Fatal(err)
	}

	// Go through each file.
	for _, f := range dir {
		// Get the name.
		name := f.Name()

		// If the name is .DS_Store, ignore it.
		if name == ".DS_Store" {
			continue
		}

		if f.IsDir() {
			// If it is a directory, process it.
			t.Run(name, func(t *testing.T) {
				processFolder(filepath.Join(fp, name), t)
			})
		} else {
			// Create a sub-test for the file.
			t.Run(name, func(t *testing.T) {
				processFile(filepath.Join(fp, name), t)
			})
		}
	}
}

func TestParse(t *testing.T) {
	utils.GlobalLock.Acquire("test:spew", func() {
		// Setup spew.
		spew.Config.DisablePointerAddresses = true
		spew.Config.SortKeys = true

		// If RESULTS_UPDATE is set to 1, we should delete the testdata/results directory.
		if os.Getenv("RESULTS_UPDATE") == "1" {
			err := os.RemoveAll(resultsReplacement)
			if err != nil {
				t.Fatal(err)
			}
		}

		// Process the folder.
		processFolder(startPrefix, t)
	})
}

//go:embed testdata/tests/misc/combo.rql
var combo string

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ast.Parse(combo)
	}
}
