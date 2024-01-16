// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package radisk

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
)

// ErrNotFound is used to define the error when the page is not found.
var ErrNotFound = errors.New("not found")

// Filesystem is used to define the interface for the filesystem.
type Filesystem interface {
	// GetPage is used to get a page from the filesystem. Returns ErrNotFound if the page does not exist.
	GetPage(page uint64) ([]byte, error)

	// SetPage is used to set a page in the filesystem.
	SetPage(page uint64, data []byte) error
}

// FSImpl is used to define the implementation of the filesystem.
type FSImpl struct {
	// Path is the path to the KV store.
	Path string
}

func (fs FSImpl) GetPage(page uint64) ([]byte, error) {
	fp := filepath.Join(fs.Path, "p_"+strconv.FormatUint(page, 10))
	f, err := os.ReadFile(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return f, nil
}

func (fs FSImpl) SetPage(page uint64, data []byte) error {
	fp := filepath.Join(fs.Path, "p_"+strconv.FormatUint(page, 10))
	return os.WriteFile(fp, data, 0644)
}

var _ Filesystem = FSImpl{}
