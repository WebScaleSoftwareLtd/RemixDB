// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package localfs

import (
	"os"
	"path/filepath"

	"github.com/juju/fslock"
	"remixdb.io/internal/engine"
	"remixdb.io/internal/engine/localfs/acid"
	"remixdb.io/internal/engine/localfs/session"
	"remixdb.io/internal/logger"
)

type Engine struct {
	c credentialsCache
	l partitionLocks
	s session.Cache

	path   string
	logger logger.Logger
}

func (e *Engine) CreateSession(partition string) (engine.Session, error) {
	// Ensure the partition stays alive until the end of the session.
	unlock, _, err := e.usePartition(partition, false)
	if err != nil {
		return nil, err
	}

	// Return the session.
	return &session.Session{
		Logger:        e.logger,
		Transaction:   acid.New(e.path),
		PartitionName: partition,
		Cache:         &e.s,
		DataFolder:    e.path,
		RelativePath:  e.getPartitionPath(partition, true),
		Unlocker:      unlock,
	}, nil
}

func (e *Engine) CreateSchemaWriteSession(partition string) (engine.Session, error) {
	// Ensure the partition stays alive until the end of the session.
	unlock, _, err := e.usePartition(partition, true)
	if err != nil {
		return nil, err
	}

	// Return the session.
	return &session.Session{
		Logger:          e.logger,
		Cache:           &e.s,
		Transaction:     acid.New(e.path),
		PartitionName:   partition,
		DataFolder:      e.path,
		RelativePath:    e.getPartitionPath(partition, true),
		SchemaWriteLock: true,
		Unlocker:        unlock,
	}, nil
}

var _ engine.Engine = (*Engine)(nil)

// New is used to create a new engine. If path is empty, the environment variable REMIXDB_DATA_PATH is used or
// ~/.remixdb/data if it is not set.
func New(logger logger.Logger, path string) engine.Engine {
	// Tag the logger.
	logger = logger.Tag("engine.localfs")

	// Handles the default path.
	if path == "" {
		path = os.Getenv("REMIXDB_DATA_PATH")
		if path == "" {
			homedir, err := os.UserHomeDir()
			if err != nil {
				panic(err)
			}
			path = filepath.Join(homedir, ".remixdb", "data")
		}
	}

	// Make the directory if it does not exist.
	err := os.MkdirAll(filepath.Join(path, "partitions"), 0755)
	if err != nil {
		panic(err)
	}

	// Attempt to grab the filesystem lock or exit.
	err = fslock.New(filepath.Join(path, "lock")).TryLock()
	if err != nil {
		logger.Error("The database engine storage is already locked", nil)
		os.Exit(1)
	}

	// Perform a integrity check on the database.
	integrityCheck(path)

	// Return the engine.
	return &Engine{path: path, logger: logger}
}
