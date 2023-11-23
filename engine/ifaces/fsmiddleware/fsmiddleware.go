// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package fsmiddleware

// Middleware is used to define the filesystem middleware interface.
type Middleware interface {
	// ReadFile is used to try and read a file from the cache. If it can't, it will invoke the next
	// middleware and then possibly insert it into the cache.
	ReadFile(rel, partition string, partitionTtl uint64, next func() ([]byte, error)) ([]byte, error)

	// DeleteFile is used to delete a file from the cache.
	DeleteFile(rel, partition string) error

	// DeletePartition is used to delete a partition from the cache.
	DeletePartition(rel, partition string) error

	// WriteFile is used to write a file to the cache.
	WriteFile(rel, partition string, partitionTtl uint64, b []byte) error

	// RenameFile is used to rename a file in the cache.
	RenameFile(oldRel, newRel, partition string, partitionTtl uint64) error
}
