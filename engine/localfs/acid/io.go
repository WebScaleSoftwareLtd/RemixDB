// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package acid

import (
	"io/fs"
	"os"
)

// DeleteAllSafely will in anything under the engines filesystem, this will ensure that deletes are always properly ran on
// the next boot if they do not happen as expected.
func DeleteAllSafely(path string) error {
	// Write a .RM file with no contents.
	err := os.WriteFile(path+".RM", []byte{}, 0644)
	if err != nil {
		return err
	}
	defer os.Remove(path + ".RM")

	// Delete the file.
	err = os.RemoveAll(path)
	if err != nil {
		return err
	}

	// Return nil.
	return nil
}

// WriteSafely will in anything under the engines filesystem, this will ensure that writes are always properly ran on
// the next boot if they do not happen as expected.
func WriteSafely(path string, b []byte, perm fs.FileMode) error {
	// Write a .$ file with the contents.
	err := os.WriteFile(path+".$", b, perm)
	if err != nil {
		return err
	}

	// Write a .R file with no contents.
	err = os.WriteFile(path+".R", []byte{}, 0644)
	if err != nil {
		return err
	}
	defer os.Remove(path + ".R")

	// Remove all from the path.
	err = os.RemoveAll(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Rename the .$ file to the path.
	return os.Rename(path+".$", path)
}
