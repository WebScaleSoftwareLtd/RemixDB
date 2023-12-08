// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package localfs

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func handleFileWriteInterrupted(path string) {
	// Swap the .$ suffix for a .R suffix and see if the file exists.
	originFile := path[:len(path)-2]
	rPath := originFile + ".R"
	_, err := os.Stat(rPath)
	if err != nil {
		if os.IsNotExist(err) {
			// The power cut happened before everything was marked as ready to be
			// written. Therefore, the system was not expecting this to have been
			// written to disk. Delete the file.
			err = os.RemoveAll(path)
			if err != nil {
				panic(err)
			}
			return
		}

		// Another error occurred.
		panic(err)
	}

	// The power cut happened after everything was marked as ready to be written. Therefore,
	// do the rename after we deleted the file it is replacing.
	err = os.RemoveAll(originFile)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	err = os.Rename(path, originFile)
	if err != nil {
		panic(err)
	}

	// Delete the .R file.
	err = os.Remove(rPath)
	if err != nil {
		panic(err)
	}
}

func handleFileDeleteInterrupted(path string) {
	// Get the original file path.
	originFile := path[:len(path)-3]

	// Check if the original file exists.
	_, err := os.Stat(originFile)
	if err != nil {
		if os.IsNotExist(err) {
			// The power cut happened before the file was deleted. Therefore, the system
			// was not expecting this to have been deleted. Delete the file.
			err = os.Remove(path)
			if err != nil {
				panic(err)
			}
			return
		}

		// Another error occurred.
		panic(err)
	}

	// Remove the original file.
	err = os.RemoveAll(originFile)
	if err != nil {
		panic(err)
	}

	// Remove the .RM file.
	err = os.Remove(path)
	if err != nil {
		panic(err)
	}
}

func integrityCheck(path string) {
	dir, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	otherDirs := make([]string, 0, len(dir))
	for _, d := range dir {
		if d.IsDir() {
			// Handle directories later.
			otherDirs = append(otherDirs, d.Name())
			continue
		}

		name := d.Name()
		switch {
		case strings.HasSuffix(name, ".$"):
			handleFileWriteInterrupted(filepath.Join(path, name))
		case strings.HasSuffix(name, ".RM"):
			handleFileDeleteInterrupted(filepath.Join(path, name))
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(len(otherDirs))
	for _, d := range otherDirs {
		go func(d string) {
			defer wg.Done()
			integrityCheck(filepath.Join(path, d))
		}(d)
	}
	wg.Wait()
}
