//go:build ignore

// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package main

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/matryer/moq/pkg/moq"
)

var cfg = moq.Config{
	SrcDir:  "../engine",
	PkgName: "mocksession",
}

func main() {
	// Create the mock generator.
	m, err := moq.New(cfg)
	if err != nil {
		panic(err)
	}
	var b bytes.Buffer
	if err = m.Mock(&b, "Session"); err != nil {
		panic(err)
	}

	// Write the file to disk.
	if err = os.Mkdir("mocksession", 0755); err != nil {
		if !os.IsExist(err) {
			panic(err)
		}
	}
	if err = os.WriteFile(filepath.Join("mocksession", "mock_session_gen.go"), b.Bytes(), 0644); err != nil {
		panic(err)
	}
}
