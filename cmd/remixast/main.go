// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package main

import (
	"os"

	"github.com/davecgh/go-spew/spew"
	"remixdb.io/ast"
)

func main() {
	// Get the first argument from the command line.
	if len(os.Args) < 2 {
		panic("No file path provided.")
	}
	filePath := os.Args[1]

	// Parse the file.
	b, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	file := string(b)
	parsed, perr := ast.Parse(file)
	if perr != nil {
		_, _ = os.Stderr.WriteString("Error parsing file:\n")
		spew.Fdump(os.Stderr, perr)
		os.Exit(1)
	}
	spew.Dump(parsed)
}
