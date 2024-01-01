//go:build windows

// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package precache

import (
	"errors"

	"github.com/urfave/cli/v2"
)

const noWindowsSupport = "Windows is not supported for the database currently."

// Precache is used to tell Windows users they can't use Windows to run the database.
func Precache(_ *cli.Context) error {
	return errors.New(noWindowsSupport)
}
