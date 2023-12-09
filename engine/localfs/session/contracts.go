// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package session

import (
	"os"
	"path/filepath"

	"github.com/vmihailenco/msgpack/v5"
	"remixdb.io/ast"
	"remixdb.io/engine"
)

func (s *Session) loadContracts() (map[string]*ast.ContractToken, error) {
	contracts, ok := s.Cache.contracts.Get(s.PartitionName)
	if !ok {
		// Load the contracts file from disk.
		b, err := os.ReadFile(filepath.Join(s.Path, "contracts"))
		if err != nil {
			if os.IsNotExist(err) {
				// Return an empty map.
				return map[string]*ast.ContractToken{}, nil
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

func (s *Session) ensureWriteLock() error {
	if !s.SchemaWriteLock {
		return engine.ErrReadOnlySession
	}
	return nil
}

func (s *Session) writeContractTombstone(contract *ast.ContractToken) error {
	// Read the tombstones file.
	sl := []*ast.ContractToken{}
	b, err := os.ReadFile(filepath.Join(s.Path, "contract_tombstones"))
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
	if err := s.ensureWriteLock(); err != nil {
		return err
	}

	contracts, err := s.loadContracts()
	if err != nil {
		return err
	}

	c, ok := contracts[key]
	if !ok {
		return engine.ErrNotExists
	}
	s.Cache.contracts.Delete(s.PartitionName)
	delete(contracts, key)

	b, err := msgpack.Marshal(contracts)
	if err != nil {
		return err
	}
	s.Transaction.WriteFile(filepath.Join(s.RelativePath, "contracts"), b)
	s.writeContractTombstone(c)
	return nil
}

func (s *Session) WriteContract(contract *ast.ContractToken) error {
	if err := s.ensureWriteLock(); err != nil {
		return err
	}

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
	contracts[contract.Name] = contract
	b, err := msgpack.Marshal(contracts)
	if err != nil {
		return err
	}

	s.Transaction.WriteFile(filepath.Join(s.RelativePath, "contracts"), b)
	return nil
}

func (s *Session) Contracts() (contracts []*ast.ContractToken, err error) {
	contractsMap, err := s.loadContracts()
	if err != nil {
		return nil, err
	}

	contracts = make([]*ast.ContractToken, len(contractsMap))
	i := 0
	for _, v := range contractsMap {
		contracts[i] = v
		i++
	}
	return
}

func (s *Session) ContractTombstones() (contracts []*ast.ContractToken, err error) {
	sl := []*ast.ContractToken{}
	b, err := os.ReadFile(filepath.Join(s.Path, "contract_tombstones"))
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
