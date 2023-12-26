// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package localfs

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"sync"

	"remixdb.io/internal/engine"
)

type partitionLocks struct {
	mu sync.Mutex
	m  map[string]*sync.RWMutex
}

func (e *Engine) getPartitionLock(partition string) *sync.RWMutex {
	e.l.mu.Lock()
	defer e.l.mu.Unlock()

	if e.l.m == nil {
		e.l.m = map[string]*sync.RWMutex{}
	}

	if e.l.m[partition] == nil {
		e.l.m[partition] = &sync.RWMutex{}
	}

	return e.l.m[partition]
}

func (e *Engine) getPartitionPath(partition string, rel bool) string {
	enc := base64.URLEncoding.EncodeToString([]byte(partition))
	if rel {
		return filepath.Join("partitions", enc)
	}
	return filepath.Join(e.path, "partitions", enc)
}

func (e *Engine) CreatePartition(partition string) error {
	mu := e.getPartitionLock(partition)
	mu.Lock()
	defer mu.Unlock()

	err := os.Mkdir(e.getPartitionPath(partition, false), 0755)
	if err != nil {
		if os.IsExist(err) {
			return engine.ErrPartitionAlreadyExists
		}

		return err
	}

	return nil
}

func (e *Engine) DeletePartition(partition string) error {
	mu := e.getPartitionLock(partition)
	mu.Lock()
	defer mu.Unlock()

	err := os.RemoveAll(e.getPartitionPath(partition, false))
	if err != nil {
		if os.IsNotExist(err) {
			return engine.ErrPartitionDoesNotExist
		}

		return err
	}
	e.c.removePartition(partition)
	e.s.CleanPartition(partition)

	return nil
}

func (e *Engine) usePartition(partition string, write bool) (unlocker func(), path string, err error) {
	mu := e.getPartitionLock(partition)
	if write {
		mu.Lock()
	} else {
		mu.RLock()
	}

	path = e.getPartitionPath(partition, false)
	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = engine.ErrPartitionDoesNotExist
		}

		if write {
			mu.Unlock()
		} else {
			mu.RUnlock()
		}
		return
	}

	unlocker = mu.RUnlock
	if write {
		unlocker = mu.Unlock
	}
	return
}

func (e *Engine) Partitions() []string {
	// List the partitions folder.
	fp := filepath.Join(e.path, "partitions")
	dir, err := os.ReadDir(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}
		}

		panic(err)
	}

	// Go through each partition.
	partitions := make([]string, 0, len(dir))
	for _, d := range dir {
		if d.IsDir() {
			// Decode the partition name.
			dec, err := base64.URLEncoding.DecodeString(d.Name())
			if err != nil {
				panic(err)
			}

			// Add the partition name.
			partitions = append(partitions, string(dec))
		}
	}
	return partitions
}
