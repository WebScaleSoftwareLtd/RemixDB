// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package session

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"

	"github.com/vmihailenco/msgpack/v5"
	"remixdb.io/ast"
	"remixdb.io/engine"
	"remixdb.io/utils"
)

func (s *Session) loadStructs() (map[string]possibleRename, error) {
	// Load the structs for this partition.
	structs, ok := s.Cache.structs.Get(s.PartitionName)
	if ok {
		// This is a cache hit. Return early.
		return structs, nil
	}

	// Read the structs file.
	b, err := s.Transaction.ReadFile(filepath.Join(s.RelativePath, "structs"))
	if err != nil {
		if os.IsNotExist(err) {
			// Return ErrNotExists.
			return nil, engine.ErrNotExists
		}

		return nil, err
	}

	// Decompress the structs file.
	r, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	b, err = io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if err := r.Close(); err != nil {
		return nil, err
	}

	// Unmarshal the structs file.
	err = msgpack.Unmarshal(b, &structs)
	if err != nil {
		return nil, err
	}

	// Cache the structs file.
	s.Cache.structs.Set(s.PartitionName, structs)

	// Return the structs.
	return structs, nil
}

func (s *Session) GetStructByKey(key string) (structHistory []*ast.StructToken, err error) {
	// Load the structs for this partition.
	structs, err := s.loadStructs()
	if err != nil {
		return nil, err
	}

	// Get the struct metadata.
	v := structs[key]

handleStruct:
	// Handle if this was a rename.
	if v.R != nil {
		v = structs[*v.R]
		goto handleStruct
	}

	structHistory = v.S
	if structHistory == nil {
		err = engine.ErrNotExists
	}
	return
}

func (s *Session) writeStructTombstone(structHistory []*ast.StructToken) error {
	// Read the tombstones file.
	m := map[string]possibleRename{}
	fp := filepath.Join(s.RelativePath, "struct_tombstones")
	b, err := s.Transaction.ReadFile(fp)
	if err == nil {
		// Decompress the tombstones file.
		r, err := gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			return err
		}
		b, err = io.ReadAll(r)
		if err != nil {
			return err
		}
		if err := r.Close(); err != nil {
			return err
		}

		// Unmarshal the tombstones file.
		err = msgpack.Unmarshal(b, &m)
		if err != nil {
			return err
		}
	} else {
		// If this isn't a not exists error, return it.
		if !os.IsNotExist(err) {
			return err
		}
	}

	// Get all previous names.
	names := map[string]struct{}{}
	for _, v := range structHistory {
		names[v.Name] = struct{}{}
	}
	newName := structHistory[len(structHistory)-1].Name
	delete(names, newName)

	// Mark all the renames in tombstones.
	for name := range names {
		m[name] = possibleRename{
			R: &newName,
		}
	}

	// Set the final struct.
	m[newName] = possibleRename{
		S: structHistory,
	}

	// Marshal the tombstones file.
	b, err = msgpack.Marshal(m)
	if err != nil {
		return err
	}

	// Compress the tombstones file.
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, err = w.Write(b)
	if err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	b = buf.Bytes()

	// Write the tombstones file.
	s.Transaction.WriteFile(fp, b)

	// No errors!
	return nil
}

func (s *Session) DeleteStructByKey(key string) error {
	// Ensure the session has a write lock.
	if err := s.ensureWriteLock(); err != nil {
		return err
	}

	// Load the structs for this partition.
	structs, err := s.loadStructs()
	if err != nil {
		return err
	}

	// Get the struct metadata.
	v, ok := structs[key]
	if !ok {
		return engine.ErrNotExists
	}

	// If this is a rename, return not exists.
	if v.R != nil {
		return engine.ErrNotExists
	}

	// Find all previous names for the struct.
	structHistory := v.S
	names := map[string]struct{}{}
	for _, s := range structHistory {
		names[s.Name] = struct{}{}
	}

	// Drop from the cache.
	s.Cache.structs.Delete(s.PartitionName)

	// Kill all names related to this struct.
	for name := range names {
		delete(structs, name)
	}

	// Marshal the structs contents.
	b, err := msgpack.Marshal(structs)
	if err != nil {
		return err
	}

	// Compress the structs contents.
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, err = w.Write(b)
	if err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	b = buf.Bytes()

	// Journal the structs file and tombstones file.
	s.Transaction.WriteFile(filepath.Join(s.RelativePath, "structs"), b)
	s.writeStructTombstone(structHistory)

	// Journal deleting the struct folder.
	s.Transaction.DeleteAll(filepath.Join(s.RelativePath, "tables", base64.URLEncoding.EncodeToString([]byte(key))))

	// Return no errors.
	return nil
}

func (s *Session) Structs() (structs []*ast.StructToken, err error) {
	structsMap, err := s.loadStructs()
	if err != nil {
		if err == engine.ErrNotExists {
			// Set the map and jump.
			structsMap = map[string]possibleRename{}
			goto postError
		}
		return nil, err
	}

postError:
	structs = make([]*ast.StructToken, len(structsMap))
	i := 0
	for _, v := range structsMap {
		if v.R != nil {
			continue
		}
		structs[i] = v.S[len(v.S)-1]
		i++
	}
	return structs, nil
}

func (s *Session) StructTombstones() (renames map[string]string, structs []*ast.StructToken, err error) {
	// Load the tombstones file.
	m := map[string]possibleRename{}
	fp := filepath.Join(s.RelativePath, "struct_tombstones")
	b, err := s.Transaction.ReadFile(fp)
	if err == nil {
		// Decompress the tombstones file.
		r, err := gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, nil, err
		}
		b, err = io.ReadAll(r)
		if err != nil {
			return nil, nil, err
		}
		if err := r.Close(); err != nil {
			return nil, nil, err
		}

		// Unmarshal the tombstones file.
		err = msgpack.Unmarshal(b, &m)
		if err != nil {
			return nil, nil, err
		}
	} else {
		// If this isn't a not exists error, return it.
		if !os.IsNotExist(err) {
			return nil, nil, err
		}
	}

	// Make the map for the renames.
	renames = map[string]string{}

	// Make sure the slice isn't nil.
	structs = make([]*ast.StructToken, 0, len(m))

	// Go through and map everything as expected.
	for k, v := range m {
		if v.R == nil {
			structs = append(structs, v.S[len(v.S)-1])
		} else {
			renames[k] = *v.R
		}
	}

	// Return no errors.
	return renames, structs, nil
}

func (s *Session) getPartitionObjectLockMutex(name string) *utils.NamedLock {
	// Start with a read lock since we are hopeful we can find fast.
	s.Cache.objectLocksMu.RLock()

	// Try getting from the cache as-is.
	var l *utils.NamedLock
	var ok bool
	if s.Cache.objectLocks != nil {
		l, ok = s.Cache.objectLocks[s.PartitionName]
	}

	// Now read unlock.
	s.Cache.objectLocksMu.RUnlock()

	if !ok {
		// Since the partition doesn't even exist, we need a write lock on global.
		s.Cache.objectLocksMu.Lock()

		// Make sure objectLocks even exists.
		if s.Cache.objectLocks == nil {
			s.Cache.objectLocks = map[string]*utils.NamedLock{}
		}

		// Try again in case we got raced.
		l, ok = s.Cache.objectLocks[s.PartitionName]
		if !ok {
			// Ok, we didn't. Make the lock.
			l = &utils.NamedLock{}
			s.Cache.objectLocks[s.PartitionName] = l
		}

		// Unlock the global map.
		s.Cache.objectLocksMu.Unlock()
	}

	// Return the lock.
	return l
}

func (s *Session) AcquireStructObjectWriteLock(structName string, keys ...[]byte) error {
	// Get the object locker for this partition.
	l := s.getPartitionObjectLockMutex(structName)

	// Acquire the locks.
	for _, key := range keys {
		l.Lock(structName + " " + string(key))
	}

	// Return no errors since this can't error locally.
	return nil
}

func (s *Session) ReleaseStructObjectWriteLock(structName string, keys ...[]byte) error {
	// Get the object locker for this partition.
	l := s.getPartitionObjectLockMutex(structName)

	// Release the locks.
	for _, key := range keys {
		l.Unlock(structName + " " + string(key))
	}

	// Return no errors since this can't error locally.
	return nil
}

func (s *Session) AcquireStructObjectReadLock(structName string, keys ...[]byte) error {
	// Get the object locker for this partition.
	l := s.getPartitionObjectLockMutex(structName)

	// Acquire the locks.
	for _, key := range keys {
		l.RLock(structName + " " + string(key))
	}

	// Return no errors since this can't error locally.
	return nil
}

func (s *Session) ReleaseStructObjectReadLock(structName string, keys ...[]byte) error {
	// Get the object locker for this partition.
	l := s.getPartitionObjectLockMutex(structName)

	// Release the locks.
	for _, key := range keys {
		l.RUnlock(structName + " " + string(key))
	}

	// Return no errors since this can't error locally.
	return nil
}

var _ engine.StructSessionMethods = (*Session)(nil)
