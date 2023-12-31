// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package engine

import (
	"errors"

	"remixdb.io/ast"
)

// ErrNotExists is used to define the error when the key does not exist.
var ErrNotExists = errors.New("key does not exist")

// ErrNotTable is used to define the error when the struct is not a table.
var ErrNotTable = errors.New("struct is not a table")

// StructSessionMethods is used to define the methods for the struct session.
type StructSessionMethods interface {
	// GetStructByKey is used to get the struct for a specified key. If the key does not
	// exist, the error ErrNotExists is returned. The slice repersents the history of the
	// struct, with the last item being the most recent. Note the positions on the AST tokens
	// are not set.
	GetStructByKey(key string) (structHistory []*ast.StructToken, err error)

	// DeleteStructByKey is used to delete the struct for a specified key. When a struct
	// is deleted, it is put into a tombstone state. If the key does not exist, the error
	// ErrNotExists is returned.
	DeleteStructByKey(key string) error

	// Structs is used to return all of the latest structs in the database partition. Note the
	// positions on the AST tokens are not set.
	Structs() (structs []*ast.StructToken, err error)

	// StructTombstones returns all structs in the state they were when they were deleted. Note the
	// positions on the AST tokens are not set.
	StructTombstones() (renames map[string]string, structs []*ast.StructToken, err error)

	// AcquireStructReadLock is used to acquire a read lock on a struct. This is used to prevent
	// concurrent writes to the same struct. The lock is released when the session is closed or
	// ReleaseStructReadLock is called. This or a write lock must be acquired before editing a
	// struct.
	AcquireStructReadLock(structNames ...string) error

	// ReleaseStructReadLock is used to release a read lock on a struct. This is used to allow for
	// efficiencies inside of a contracts compiled logic.
	ReleaseStructReadLock(structNames ...string) error

	// AcquireStructWriteLock is used to acquire a write lock on a struct. This is used to prevent
	// concurrent writes to the same struct. The lock is released when the session is closed or
	// ReleaseStructWriteLock is called.
	AcquireStructWriteLock(structNames ...string) error

	// ReleaseStructWriteLock is used to release a write lock on a struct. This is used to allow for
	// efficiencies inside of a contracts compiled logic.
	ReleaseStructWriteLock(structNames ...string) error

	// AcquireStructObjectWriteLock is used to acquire a write lock on struct object(s). This is used
	// to prevent concurrent writes to the same struct object. The lock is released when the session
	// is closed or ReleaseStructObjectWriteLock is called.
	AcquireStructObjectWriteLock(structName string, keys ...[]byte) error

	// ReleaseStructObjectWriteLock is used to release a write lock on struct object(s). This is used
	// to allow for efficiencies inside of a contracts compiled logic.
	ReleaseStructObjectWriteLock(structName string, keys ...[]byte) error

	// AcquireStructObjectReadLock is used to acquire a read lock on struct object(s). This is used
	// to prevent concurrent writes to the same struct object. The lock is released when the session
	// is closed or ReleaseStructObjectReadLock is called.
	AcquireStructObjectReadLock(structName string, keys ...[]byte) error

	// ReleaseStructObjectReadLock is used to release a read lock on struct object(s). This is used
	// to allow for efficiencies inside of a contracts compiled logic.
	ReleaseStructObjectReadLock(structName string, keys ...[]byte) error
}

// ContractSessionMethods is used to define the methods for the contract session.
type ContractSessionMethods interface {
	// GetContractByKey is used to get the contract for a specified key. If the key does not
	// exist, the error ErrNotExists is returned.
	GetContractByKey(key string) (contract *ast.ContractToken, err error)

	// DeleteContractByKey is used to delete the contract for a specified key. When a contract
	// is deleted, it is put into a tombstone state. If the key does not exist, the error
	// ErrNotExists is returned.
	DeleteContractByKey(key string) error

	// WriteContract is used to write a contract. If the contract already exists, it will be
	// overwritten.
	WriteContract(contract *ast.ContractToken) error

	// Contracts is used to get all of the contracts.
	Contracts() (contracts []*ast.ContractToken, err error)

	// ContractTombstones returns all contracts in the state they were when they were deleted.
	ContractTombstones() (contracts []*ast.ContractToken, err error)
}

// Session is used to define a session. You must call Close on the session.
type Session interface {
	// Close is used to close the session. If this is a write session, it will be rolled back if
	// it has not been committed.
	Close() error

	// Rollback is used to rollback any changes made in the session.
	Rollback() error

	// Commit is used to commit any changes made in the session. Rollback will then work for changes
	// made after this commit only.
	Commit() error

	StructSessionMethods
	ContractSessionMethods
}

// ErrPartitionDoesNotExist is used to define the error when the partition does not exist.
var ErrPartitionDoesNotExist = errors.New("partition does not exist")

// ErrPartitionAlreadyExists is used to define the error when the partition already exists.
var ErrPartitionAlreadyExists = errors.New("partition already exists")

// ErrReadOnlySession is used to define the error when a write is attempted on a read session.
var ErrReadOnlySession = errors.New("read only session")

// Engine is used to define the interface for the engine.
type Engine interface {
	// CreateReadSession is used to create a read session. You must call Close on the read session
	// when you are done with it. If the partition does not exist, the error ErrPartitionDoesNotExist
	// is returned. Read sessions cannot be used to write to the schema, any attempt to do so will
	// return a ErrReadOnlySession error.
	CreateSession(partition string) (Session, error)

	// CreateWriteSession is used to create a write session. You must call Close on the write session when
	// you are done with it. If the partition does not exist, the error ErrPartitionDoesNotExist
	// is returned.
	CreateSchemaWriteSession(partition string) (Session, error)

	// CreatePartition is used to create a partition. Returns ErrPartitionAlreadyExists if the partition
	// already exists.
	CreatePartition(partition string) error

	// DeletePartition is used to delete a partition. Returns ErrPartitionDoesNotExist if the partition does not exist.
	DeletePartition(partition string) error

	// GetAuthenticationPermissionsByAPIKey is used to get the authentication permissions for a specified API key.
	// This is generally used for authentication on load. If the slice is nil, the API key does not exist.
	GetAuthenticationPermissionsByAPIKey(partition, apiKey string) (username string, permissions []string, err error)

	// GetAuthenticationPermissionsByUsername is used to get the authentication permissions for a specified username.
	// If the slice is nil, the username does not exist.
	GetAuthenticationPermissionsByUsername(partition, username string) (permissions []string, err error)

	// SetAuthenticationPermissions is used to set the authentication permissions for a specified username. If the
	// username does not exist, it will be created.
	SetAuthenticationPermissions(partition, username string, permissions []string) error

	// Usernames is used to get the usernames for a specified partition.
	Usernames(partition string) ([]string, error)

	// CreateAPIKeyForUsername is used to create an API key for a specified username. If the username does not exist,
	// it will be created.
	CreateAPIKeyForUsername(partition, username, apiKey string) error

	// DeleteAPIKey is used to delete an API key. If the API key does not exist, it will return nil.
	DeleteAPIKey(partition, apiKey string) error

	// DeleteUsername is used to delete a username. If the username does not exist, it will return nil.
	DeleteUsername(partition, username string) error

	// Partitions is used to get all of the partitions.
	Partitions() []string
}
