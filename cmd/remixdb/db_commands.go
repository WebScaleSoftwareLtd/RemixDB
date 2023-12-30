// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package main

import (
	_ "embed"
	"strings"

	"github.com/urfave/cli/v2"
	"remixdb.io/cmd/remixdb/dbstart"
)

//go:embed dbstart/command_description.txt
var dbStartDescription string

var dbCommand = &cli.Command{
	Name:        "db",
	Description: "Commands relating to the management of a RemixDB database.",
	Subcommands: []*cli.Command{
		{
			Name:        "start",
			Usage:       "Starts the RemixDB database with values from the YAML configuration or environment variables. Not supported on Windows. See the description for more information.",
			Description: strings.TrimSpace(dbStartDescription),
			Action:      dbstart.Start,
		},
	},
}

func init() {
	app.Commands = append(app.Commands, dbCommand)
}
