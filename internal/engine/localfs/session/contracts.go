// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package session

import (
	"os"
	"path/filepath"

	"github.com/vmihailenco/msgpack/v5"
	"remixdb.io/ast"
	"remixdb.io/internal/engine"
)

func (s *Session) loadContracts() (map[string]*ast.ContractToken, error) {
	contracts, ok := s.Cache.contracts.Get(s.PartitionName)
	if !ok {
		// Load the contracts file from disk.
		b, err := s.Transaction.ReadFile(filepath.Join(s.RelativePath, "contracts"))
		if err != nil {
			if os.IsNotExist(err) {
				// Return ErrNotExists.
				return nil, engine.ErrNotExists
			}

			// This is another type of error, so return it.
			return nil, err
		}

		// Unmarshal the contracts file.
		err = msgpack.Unmarshal(b, &contracts)
		if err != nil {
			return nil, err
		}

		// Cache the contracts file.
		s.Cache.contracts.Set(s.PartitionName, contracts)
	}
	return contracts, nil
}

func (s *Session) GetContractByKey(key string) (contract *ast.ContractToken, err error) {
	contracts, err := s.loadContracts()
	if err != nil {
		return nil, err
	}

	contract = contracts[key]
	if contract == nil {
		err = engine.ErrNotExists
	}
	return
}

func (s *Session) writeContractTombstone(contract *ast.ContractToken) error {
	// Read the tombstones file.
	sl := []*ast.ContractToken{}
	b, err := s.Transaction.ReadFile(filepath.Join(s.RelativePath, "contract_tombstones"))
	if err == nil {
		// Unmarshal the tombstones file.
		err = msgpack.Unmarshal(b, &sl)
		if err != nil {
			return err
		}
	} else {
		// If this isn't a not exists error, return it.
		if !os.IsNotExist(err) {
			return err
		}
	}

	// Append the contract to the list.
	sl = append(sl, contract)

	// Marshal the tombstones file.
	b, err = msgpack.Marshal(sl)
	if err != nil {
		return err
	}

	// Write the tombstones file.
	s.Transaction.WriteFile(filepath.Join(s.RelativePath, "contract_tombstones"), b)
	return nil
}

func (s *Session) DeleteContractByKey(key string) error {
	// Ensure the session has a write lock.
	if err := s.ensureWriteLock(); err != nil {
		return err
	}

	// Load the contracts.
	contracts, err := s.loadContracts()
	if err != nil {
		return err
	}

	// Make sure the contract is present. If it is, remove it.
	c, ok := contracts[key]
	if !ok {
		return engine.ErrNotExists
	}
	s.Cache.contracts.Delete(s.PartitionName)
	delete(contracts, key)

	// Journal the action.
	b, err := msgpack.Marshal(contracts)
	if err != nil {
		return err
	}
	s.Transaction.WriteFile(filepath.Join(s.RelativePath, "contracts"), b)
	s.writeContractTombstone(c)
	return nil
}

func (s *Session) WriteContract(contract *ast.ContractToken) error {
	// Ensure the session has a write lock.
	if err := s.ensureWriteLock(); err != nil {
		return err
	}

	// Load the contracts and then drop them from the cache if they exist.
	contracts, err := s.loadContracts()
	if err != nil {
		if err == engine.ErrNotExists {
			// Small little micro-optimization.
			contracts = map[string]*ast.ContractToken{}
			goto postCacheHandling
		}
		return err
	}
	s.Cache.contracts.Delete(s.PartitionName)

postCacheHandling:
	// Journal the contracts edit.
	contracts[contract.Name] = contract
	b, err := msgpack.Marshal(contracts)
	if err != nil {
		return err
	}
	s.Transaction.WriteFile(filepath.Join(s.RelativePath, "contracts"), b)

	// Load the contract tombstones and remove the contract from it if it exists.
	tombstones, err := s.ContractTombstones()
	if err != nil {
		return err
	}
	newTombstones := make([]*ast.ContractToken, 0, len(tombstones))
	for _, v := range tombstones {
		if v.Name != contract.Name {
			newTombstones = append(newTombstones, v)
		}
	}

	// Write the new tombstones.
	b, err = msgpack.Marshal(newTombstones)
	if err != nil {
		return err
	}
	s.Transaction.WriteFile(filepath.Join(s.RelativePath, "contract_tombstones"), b)

	// No errors!
	return nil
}

func (s *Session) Contracts() (contracts []*ast.ContractToken, err error) {
	contractsMap, err := s.loadContracts()
	if err != nil {
		if err == engine.ErrNotExists {
			// Set the map and jump.
			contractsMap = map[string]*ast.ContractToken{}
			goto postError
		}
		return nil, err
	}

postError:
	contracts = make([]*ast.ContractToken, len(contractsMap))
	i := 0
	for _, v := range contractsMap {
		contracts[i] = v
		i++
	}
	return contracts, nil
}

func (s *Session) ContractTombstones() (contracts []*ast.ContractToken, err error) {
	sl := []*ast.ContractToken{}
	b, err := s.Transaction.ReadFile(filepath.Join(s.RelativePath, "contract_tombstones"))
	if err == nil {
		// Unmarshal the tombstones file.
		err = msgpack.Unmarshal(b, &sl)
		if err != nil {
			return nil, err
		}
	} else {
		// If this isn't a not exists error, return it.
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return sl, nil
}

var _ engine.ContractSessionMethods = (*Session)(nil)
