// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package main

import (
	_ "embed"
	"strings"

	"github.com/urfave/cli/v2"
	"remixdb.io/cmd/remixdb/db/config"
	"remixdb.io/cmd/remixdb/db/precache"
	"remixdb.io/cmd/remixdb/db/start"
)

//go:embed db/start/command_description.txt
var dbStartDescription string

var dbCommand = &cli.Command{
	Name:  "db",
	Usage: "Commands relating to the management of a RemixDB database.",
	Subcommands: []*cli.Command{
		{
			Name:   "precache-go-download",
			Usage:  "Precaches the Go download for the current platform.",
			Action: precache.Precache,
		},
		{
			Name:        "start",
			Usage:       "Starts the RemixDB database with values from the YAML configuration or environment variables. Not supported on Windows. See the description for more information.",
			Description: strings.TrimSpace(dbStartDescription),
			Action:      start.Start,
		},
		{
			Name:  "config",
			Usage: "Commands relating to the configuration of the RemixDB database.",
			Subcommands: []*cli.Command{
				{
					Name:   "edit",
					Usage:  "Opens a UI to edit the configuration file.",
					Action: config.Edit,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    "config-path",
							Aliases: []string{"c"},
							Usage:   "The path to the configuration file. Overrides the REMIXDB_CONFIG_PATH environment variable.",
						},
					},
				},
			},
		},
	},
}

func init() {
	app.Commands = append(app.Commands, dbCommand)
}
