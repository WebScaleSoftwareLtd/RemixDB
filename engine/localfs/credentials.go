// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package localfs

import (
	"os"
	"path/filepath"

	"github.com/vmihailenco/msgpack/v5"
	"remixdb.io/engine/localfs/acid"
	"remixdb.io/utils"
)

type partitionCredentials struct {
	U2A map[string][]string
	A2U map[string]string
	U2P map[string][]string
}

type credentialsCache struct {
	partitions utils.TLRUCache[string, partitionCredentials]
}

func (c *credentialsCache) removePartition(partition string) {
	c.partitions.Delete(partition)
}

func (c *credentialsCache) getOrCachePartition(path, partition string) (partitionCredentials, error) {
	// Check if we cached the partition.
	partitionCache, ok := c.partitions.Get(partition)
	if !ok {
		// Load the partition.
		fp := filepath.Join(path, "credentials")
		b, err := os.ReadFile(fp)
		if err == nil {
			// Handle loading the partition.
			err = msgpack.Unmarshal(b, &partitionCache)
			if err != nil {
				return partitionCredentials{}, err
			}
		} else {
			// Handle creating the partition credentials since it does not exist.
			partitionCache = partitionCredentials{
				U2A: map[string][]string{},
				A2U: map[string]string{},
				U2P: map[string][]string{},
			}
		}

		// Cache the partition.
		c.partitions.Set(partition, partitionCache)
	}

	// Return the partition.
	return partitionCache, nil
}

func (e *Engine) GetAuthenticationPermissionsByAPIKey(partition, apiKey string) (username string, permissions []string, err error) {
	// Ensure the partition stays alive until the end of the function.
	unlock, path, err := e.usePartition(partition, false)
	if err != nil {
		return "", nil, err
	}
	defer unlock()

	// Get the partition credentials.
	partitionCreds, err := e.c.getOrCachePartition(path, partition)
	if err != nil {
		return "", nil, err
	}

	// Get the username.
	var ok bool
	username, ok = partitionCreds.A2U[apiKey]
	if !ok {
		return "", nil, nil
	}

	// Get the permissions.
	permissions = partitionCreds.U2P[username]
	return
}

func (e *Engine) GetAuthenticationPermissionsByUsername(partition, username string) (permissions []string, err error) {
	// Ensure the partition stays alive until the end of the function.
	unlock, path, err := e.usePartition(partition, false)
	if err != nil {
		return nil, err
	}
	defer unlock()

	// Get the partition credentials.
	partitionCreds, err := e.c.getOrCachePartition(path, partition)
	if err != nil {
		return nil, err
	}

	// Get the permissions.
	permissions = partitionCreds.U2P[username]
	return
}

func (e *Engine) Usernames(partition string) ([]string, error) {
	// Ensure the partition stays alive until the end of the function.
	unlock, path, err := e.usePartition(partition, false)
	if err != nil {
		return nil, err
	}
	defer unlock()

	// Get the partition credentials.
	partitionCreds, err := e.c.getOrCachePartition(path, partition)
	if err != nil {
		return nil, err
	}

	// Get the usernames.
	usernames := make([]string, len(partitionCreds.U2A))
	i := 0
	for username := range partitionCreds.U2A {
		usernames[i] = username
		i++
	}
	return usernames, nil
}

func (e *Engine) SetAuthenticationPermissions(partition, username string, permissions []string) error {
	// Ensure the partition stays alive until the end of the function.
	unlock, path, err := e.usePartition(partition, true)
	if err != nil {
		return err
	}
	defer unlock()

	// Get the partition credentials.
	partitionCreds, err := e.c.getOrCachePartition(path, partition)
	if err != nil {
		return err
	}

	// Copy U2P before we mutate it so we do not mutate the cache.
	u2p := make(map[string][]string, len(partitionCreds.U2P))
	for k, v := range partitionCreds.U2P {
		u2p[k] = v
	}
	partitionCreds.U2P = u2p

	// Set the permissions.
	partitionCreds.U2P[username] = permissions

	// Save the partition credentials.
	b, err := msgpack.Marshal(partitionCreds)
	if err != nil {
		return err
	}
	err = acid.WriteSafely(filepath.Join(path, "credentials"), b, 0644)
	if err != nil {
		return err
	}

	// Delete the partition from the cache.
	e.c.removePartition(partition)

	// No errors!
	return nil
}

func (e *Engine) CreateAPIKeyForUsername(partition, username, apiKey string) error {
	// Ensure the partition stays alive until the end of the function.
	unlock, path, err := e.usePartition(partition, true)
	if err != nil {
		return err
	}
	defer unlock()

	// Get the partition credentials.
	partitionCreds, err := e.c.getOrCachePartition(path, partition)
	if err != nil {
		return err
	}

	// Copy both U2A and A2U before we mutate them so we do not mutate the cache.
	u2a := make(map[string][]string, len(partitionCreds.U2A))
	for k, v := range partitionCreds.U2A {
		u2a[k] = v
	}
	partitionCreds.U2A = u2a
	a2u := make(map[string]string, len(partitionCreds.A2U))
	for k, v := range partitionCreds.A2U {
		a2u[k] = v
	}
	partitionCreds.A2U = a2u

	// Set the API key.
	partitionCreds.U2A[username] = append(partitionCreds.U2A[username], apiKey)
	partitionCreds.A2U[apiKey] = username

	// Save the partition credentials.
	b, err := msgpack.Marshal(partitionCreds)
	if err != nil {
		return err
	}
	err = acid.WriteSafely(filepath.Join(path, "credentials"), b, 0644)
	if err != nil {
		return err
	}

	// Delete the partition from the cache.
	e.c.removePartition(partition)

	// No errors!
	return nil
}

func (e *Engine) DeleteAPIKey(partition, apiKey string) error {
	// Ensure the partition stays alive until the end of the function.
	unlock, path, err := e.usePartition(partition, true)
	if err != nil {
		return err
	}
	defer unlock()

	// Get the partition credentials.
	partitionCreds, err := e.c.getOrCachePartition(path, partition)
	if err != nil {
		return err
	}

	// Copy both U2A and A2U before we mutate them so we do not mutate the cache.
	u2a := make(map[string][]string, len(partitionCreds.U2A))
	for k, v := range partitionCreds.U2A {
		u2a[k] = v
	}
	partitionCreds.U2A = u2a
	a2u := make(map[string]string, len(partitionCreds.A2U))
	for k, v := range partitionCreds.A2U {
		a2u[k] = v
	}
	partitionCreds.A2U = a2u

	// Get the username.
	username, ok := partitionCreds.A2U[apiKey]
	if !ok {
		return nil
	}

	// Delete the API key.
	delete(partitionCreds.A2U, apiKey)
	for i, v := range partitionCreds.U2A[username] {
		if v == apiKey {
			partitionCreds.U2A[username] = append(partitionCreds.U2A[username][:i], partitionCreds.U2A[username][i+1:]...)
			break
		}
	}

	// Save the partition credentials.
	b, err := msgpack.Marshal(partitionCreds)
	if err != nil {
		return err
	}
	err = acid.WriteSafely(filepath.Join(path, "credentials"), b, 0644)
	if err != nil {
		return err
	}

	// Delete the partition from the cache.
	e.c.removePartition(partition)

	// No errors!
	return nil
}

func (e *Engine) DeleteUsername(partition, username string) error {
	// Ensure the partition stays alive until the end of the function.
	unlock, path, err := e.usePartition(partition, true)
	if err != nil {
		return err
	}
	defer unlock()

	// Get the partition credentials.
	partitionCreds, err := e.c.getOrCachePartition(path, partition)
	if err != nil {
		return err
	}

	// Copy both U2A and A2U before we mutate them so we do not mutate the cache.
	u2a := make(map[string][]string, len(partitionCreds.U2A))
	for k, v := range partitionCreds.U2A {
		u2a[k] = v
	}
	partitionCreds.U2A = u2a
	a2u := make(map[string]string, len(partitionCreds.A2U))
	for k, v := range partitionCreds.A2U {
		a2u[k] = v
	}
	partitionCreds.A2U = a2u

	// Delete the username.
	apiKeys, ok := partitionCreds.U2A[username]
	if !ok {
		return nil
	}

	// Delete the API keys.
	for _, apiKey := range apiKeys {
		delete(partitionCreds.A2U, apiKey)
	}

	// Delete the username.
	delete(partitionCreds.U2A, username)

	// Save the partition credentials.
	b, err := msgpack.Marshal(partitionCreds)
	if err != nil {
		return err
	}
	err = acid.WriteSafely(filepath.Join(path, "credentials"), b, 0644)
	if err != nil {
		return err
	}

	// Delete the partition from the cache.
	e.c.removePartition(partition)

	// No errors!
	return nil
}
