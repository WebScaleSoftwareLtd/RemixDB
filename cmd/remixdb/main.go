// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

var app = cli.App{
	Name:  "remixdb",
	Usage: "A functional database for the modern web.",
}

func main() {
	if err := app.Run(os.Args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
