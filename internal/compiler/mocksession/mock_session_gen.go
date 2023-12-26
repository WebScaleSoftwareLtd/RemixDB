// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocksession

import (
	"remixdb.io/ast"
	"remixdb.io/internal/engine"
	"sync"
)

// Ensure, that SessionMock does implement engine.Session.
// If this is not the case, regenerate this file with moq.
var _ engine.Session = &SessionMock{}

// SessionMock is a mock implementation of engine.Session.
//
//	func TestSomethingThatUsesSession(t *testing.T) {
//
//		// make and configure a mocked engine.Session
//		mockedSession := &SessionMock{
//			AcquireStructObjectReadLockFunc: func(structName string, keys ...[]byte) error {
//				panic("mock out the AcquireStructObjectReadLock method")
//			},
//			AcquireStructObjectWriteLockFunc: func(structName string, keys ...[]byte) error {
//				panic("mock out the AcquireStructObjectWriteLock method")
//			},
//			CloseFunc: func() error {
//				panic("mock out the Close method")
//			},
//			CommitFunc: func() error {
//				panic("mock out the Commit method")
//			},
//			ContractTombstonesFunc: func() ([]*ast.ContractToken, error) {
//				panic("mock out the ContractTombstones method")
//			},
//			ContractsFunc: func() ([]*ast.ContractToken, error) {
//				panic("mock out the Contracts method")
//			},
//			DeleteContractByKeyFunc: func(key string) error {
//				panic("mock out the DeleteContractByKey method")
//			},
//			DeleteStructByKeyFunc: func(key string) error {
//				panic("mock out the DeleteStructByKey method")
//			},
//			GetContractByKeyFunc: func(key string) (*ast.ContractToken, error) {
//				panic("mock out the GetContractByKey method")
//			},
//			GetStructByKeyFunc: func(key string) ([]*ast.StructToken, error) {
//				panic("mock out the GetStructByKey method")
//			},
//			ReleaseStructObjectReadLockFunc: func(structName string, keys ...[]byte) error {
//				panic("mock out the ReleaseStructObjectReadLock method")
//			},
//			ReleaseStructObjectWriteLockFunc: func(structName string, keys ...[]byte) error {
//				panic("mock out the ReleaseStructObjectWriteLock method")
//			},
//			RollbackFunc: func() error {
//				panic("mock out the Rollback method")
//			},
//			StructTombstonesFunc: func() (map[string]string, []*ast.StructToken, error) {
//				panic("mock out the StructTombstones method")
//			},
//			StructsFunc: func() ([]*ast.StructToken, error) {
//				panic("mock out the Structs method")
//			},
//			WriteContractFunc: func(contract *ast.ContractToken) error {
//				panic("mock out the WriteContract method")
//			},
//		}
//
//		// use mockedSession in code that requires engine.Session
//		// and then make assertions.
//
//	}
type SessionMock struct {
	// AcquireStructObjectReadLockFunc mocks the AcquireStructObjectReadLock method.
	AcquireStructObjectReadLockFunc func(structName string, keys ...[]byte) error

	// AcquireStructObjectWriteLockFunc mocks the AcquireStructObjectWriteLock method.
	AcquireStructObjectWriteLockFunc func(structName string, keys ...[]byte) error

	// CloseFunc mocks the Close method.
	CloseFunc func() error

	// CommitFunc mocks the Commit method.
	CommitFunc func() error

	// ContractTombstonesFunc mocks the ContractTombstones method.
	ContractTombstonesFunc func() ([]*ast.ContractToken, error)

	// ContractsFunc mocks the Contracts method.
	ContractsFunc func() ([]*ast.ContractToken, error)

	// DeleteContractByKeyFunc mocks the DeleteContractByKey method.
	DeleteContractByKeyFunc func(key string) error

	// DeleteStructByKeyFunc mocks the DeleteStructByKey method.
	DeleteStructByKeyFunc func(key string) error

	// GetContractByKeyFunc mocks the GetContractByKey method.
	GetContractByKeyFunc func(key string) (*ast.ContractToken, error)

	// GetStructByKeyFunc mocks the GetStructByKey method.
	GetStructByKeyFunc func(key string) ([]*ast.StructToken, error)

	// ReleaseStructObjectReadLockFunc mocks the ReleaseStructObjectReadLock method.
	ReleaseStructObjectReadLockFunc func(structName string, keys ...[]byte) error

	// ReleaseStructObjectWriteLockFunc mocks the ReleaseStructObjectWriteLock method.
	ReleaseStructObjectWriteLockFunc func(structName string, keys ...[]byte) error

	// RollbackFunc mocks the Rollback method.
	RollbackFunc func() error

	// StructTombstonesFunc mocks the StructTombstones method.
	StructTombstonesFunc func() (map[string]string, []*ast.StructToken, error)

	// StructsFunc mocks the Structs method.
	StructsFunc func() ([]*ast.StructToken, error)

	// WriteContractFunc mocks the WriteContract method.
	WriteContractFunc func(contract *ast.ContractToken) error

	// calls tracks calls to the methods.
	calls struct {
		// AcquireStructObjectReadLock holds details about calls to the AcquireStructObjectReadLock method.
		AcquireStructObjectReadLock []struct {
			// StructName is the structName argument value.
			StructName string
			// Keys is the keys argument value.
			Keys [][]byte
		}
		// AcquireStructObjectWriteLock holds details about calls to the AcquireStructObjectWriteLock method.
		AcquireStructObjectWriteLock []struct {
			// StructName is the structName argument value.
			StructName string
			// Keys is the keys argument value.
			Keys [][]byte
		}
		// Close holds details about calls to the Close method.
		Close []struct {
		}
		// Commit holds details about calls to the Commit method.
		Commit []struct {
		}
		// ContractTombstones holds details about calls to the ContractTombstones method.
		ContractTombstones []struct {
		}
		// Contracts holds details about calls to the Contracts method.
		Contracts []struct {
		}
		// DeleteContractByKey holds details about calls to the DeleteContractByKey method.
		DeleteContractByKey []struct {
			// Key is the key argument value.
			Key string
		}
		// DeleteStructByKey holds details about calls to the DeleteStructByKey method.
		DeleteStructByKey []struct {
			// Key is the key argument value.
			Key string
		}
		// GetContractByKey holds details about calls to the GetContractByKey method.
		GetContractByKey []struct {
			// Key is the key argument value.
			Key string
		}
		// GetStructByKey holds details about calls to the GetStructByKey method.
		GetStructByKey []struct {
			// Key is the key argument value.
			Key string
		}
		// ReleaseStructObjectReadLock holds details about calls to the ReleaseStructObjectReadLock method.
		ReleaseStructObjectReadLock []struct {
			// StructName is the structName argument value.
			StructName string
			// Keys is the keys argument value.
			Keys [][]byte
		}
		// ReleaseStructObjectWriteLock holds details about calls to the ReleaseStructObjectWriteLock method.
		ReleaseStructObjectWriteLock []struct {
			// StructName is the structName argument value.
			StructName string
			// Keys is the keys argument value.
			Keys [][]byte
		}
		// Rollback holds details about calls to the Rollback method.
		Rollback []struct {
		}
		// StructTombstones holds details about calls to the StructTombstones method.
		StructTombstones []struct {
		}
		// Structs holds details about calls to the Structs method.
		Structs []struct {
		}
		// WriteContract holds details about calls to the WriteContract method.
		WriteContract []struct {
			// Contract is the contract argument value.
			Contract *ast.ContractToken
		}
	}
	lockAcquireStructObjectReadLock  sync.RWMutex
	lockAcquireStructObjectWriteLock sync.RWMutex
	lockClose                        sync.RWMutex
	lockCommit                       sync.RWMutex
	lockContractTombstones           sync.RWMutex
	lockContracts                    sync.RWMutex
	lockDeleteContractByKey          sync.RWMutex
	lockDeleteStructByKey            sync.RWMutex
	lockGetContractByKey             sync.RWMutex
	lockGetStructByKey               sync.RWMutex
	lockReleaseStructObjectReadLock  sync.RWMutex
	lockReleaseStructObjectWriteLock sync.RWMutex
	lockRollback                     sync.RWMutex
	lockStructTombstones             sync.RWMutex
	lockStructs                      sync.RWMutex
	lockWriteContract                sync.RWMutex
}

// AcquireStructObjectReadLock calls AcquireStructObjectReadLockFunc.
func (mock *SessionMock) AcquireStructObjectReadLock(structName string, keys ...[]byte) error {
	if mock.AcquireStructObjectReadLockFunc == nil {
		panic("SessionMock.AcquireStructObjectReadLockFunc: method is nil but Session.AcquireStructObjectReadLock was just called")
	}
	callInfo := struct {
		StructName string
		Keys       [][]byte
	}{
		StructName: structName,
		Keys:       keys,
	}
	mock.lockAcquireStructObjectReadLock.Lock()
	mock.calls.AcquireStructObjectReadLock = append(mock.calls.AcquireStructObjectReadLock, callInfo)
	mock.lockAcquireStructObjectReadLock.Unlock()
	return mock.AcquireStructObjectReadLockFunc(structName, keys...)
}

// AcquireStructObjectReadLockCalls gets all the calls that were made to AcquireStructObjectReadLock.
// Check the length with:
//
//	len(mockedSession.AcquireStructObjectReadLockCalls())
func (mock *SessionMock) AcquireStructObjectReadLockCalls() []struct {
	StructName string
	Keys       [][]byte
} {
	var calls []struct {
		StructName string
		Keys       [][]byte
	}
	mock.lockAcquireStructObjectReadLock.RLock()
	calls = mock.calls.AcquireStructObjectReadLock
	mock.lockAcquireStructObjectReadLock.RUnlock()
	return calls
}

// AcquireStructObjectWriteLock calls AcquireStructObjectWriteLockFunc.
func (mock *SessionMock) AcquireStructObjectWriteLock(structName string, keys ...[]byte) error {
	if mock.AcquireStructObjectWriteLockFunc == nil {
		panic("SessionMock.AcquireStructObjectWriteLockFunc: method is nil but Session.AcquireStructObjectWriteLock was just called")
	}
	callInfo := struct {
		StructName string
		Keys       [][]byte
	}{
		StructName: structName,
		Keys:       keys,
	}
	mock.lockAcquireStructObjectWriteLock.Lock()
	mock.calls.AcquireStructObjectWriteLock = append(mock.calls.AcquireStructObjectWriteLock, callInfo)
	mock.lockAcquireStructObjectWriteLock.Unlock()
	return mock.AcquireStructObjectWriteLockFunc(structName, keys...)
}

// AcquireStructObjectWriteLockCalls gets all the calls that were made to AcquireStructObjectWriteLock.
// Check the length with:
//
//	len(mockedSession.AcquireStructObjectWriteLockCalls())
func (mock *SessionMock) AcquireStructObjectWriteLockCalls() []struct {
	StructName string
	Keys       [][]byte
} {
	var calls []struct {
		StructName string
		Keys       [][]byte
	}
	mock.lockAcquireStructObjectWriteLock.RLock()
	calls = mock.calls.AcquireStructObjectWriteLock
	mock.lockAcquireStructObjectWriteLock.RUnlock()
	return calls
}

// Close calls CloseFunc.
func (mock *SessionMock) Close() error {
	if mock.CloseFunc == nil {
		panic("SessionMock.CloseFunc: method is nil but Session.Close was just called")
	}
	callInfo := struct {
	}{}
	mock.lockClose.Lock()
	mock.calls.Close = append(mock.calls.Close, callInfo)
	mock.lockClose.Unlock()
	return mock.CloseFunc()
}

// CloseCalls gets all the calls that were made to Close.
// Check the length with:
//
//	len(mockedSession.CloseCalls())
func (mock *SessionMock) CloseCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockClose.RLock()
	calls = mock.calls.Close
	mock.lockClose.RUnlock()
	return calls
}

// Commit calls CommitFunc.
func (mock *SessionMock) Commit() error {
	if mock.CommitFunc == nil {
		panic("SessionMock.CommitFunc: method is nil but Session.Commit was just called")
	}
	callInfo := struct {
	}{}
	mock.lockCommit.Lock()
	mock.calls.Commit = append(mock.calls.Commit, callInfo)
	mock.lockCommit.Unlock()
	return mock.CommitFunc()
}

// CommitCalls gets all the calls that were made to Commit.
// Check the length with:
//
//	len(mockedSession.CommitCalls())
func (mock *SessionMock) CommitCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockCommit.RLock()
	calls = mock.calls.Commit
	mock.lockCommit.RUnlock()
	return calls
}

// ContractTombstones calls ContractTombstonesFunc.
func (mock *SessionMock) ContractTombstones() ([]*ast.ContractToken, error) {
	if mock.ContractTombstonesFunc == nil {
		panic("SessionMock.ContractTombstonesFunc: method is nil but Session.ContractTombstones was just called")
	}
	callInfo := struct {
	}{}
	mock.lockContractTombstones.Lock()
	mock.calls.ContractTombstones = append(mock.calls.ContractTombstones, callInfo)
	mock.lockContractTombstones.Unlock()
	return mock.ContractTombstonesFunc()
}

// ContractTombstonesCalls gets all the calls that were made to ContractTombstones.
// Check the length with:
//
//	len(mockedSession.ContractTombstonesCalls())
func (mock *SessionMock) ContractTombstonesCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockContractTombstones.RLock()
	calls = mock.calls.ContractTombstones
	mock.lockContractTombstones.RUnlock()
	return calls
}

// Contracts calls ContractsFunc.
func (mock *SessionMock) Contracts() ([]*ast.ContractToken, error) {
	if mock.ContractsFunc == nil {
		panic("SessionMock.ContractsFunc: method is nil but Session.Contracts was just called")
	}
	callInfo := struct {
	}{}
	mock.lockContracts.Lock()
	mock.calls.Contracts = append(mock.calls.Contracts, callInfo)
	mock.lockContracts.Unlock()
	return mock.ContractsFunc()
}

// ContractsCalls gets all the calls that were made to Contracts.
// Check the length with:
//
//	len(mockedSession.ContractsCalls())
func (mock *SessionMock) ContractsCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockContracts.RLock()
	calls = mock.calls.Contracts
	mock.lockContracts.RUnlock()
	return calls
}

// DeleteContractByKey calls DeleteContractByKeyFunc.
func (mock *SessionMock) DeleteContractByKey(key string) error {
	if mock.DeleteContractByKeyFunc == nil {
		panic("SessionMock.DeleteContractByKeyFunc: method is nil but Session.DeleteContractByKey was just called")
	}
	callInfo := struct {
		Key string
	}{
		Key: key,
	}
	mock.lockDeleteContractByKey.Lock()
	mock.calls.DeleteContractByKey = append(mock.calls.DeleteContractByKey, callInfo)
	mock.lockDeleteContractByKey.Unlock()
	return mock.DeleteContractByKeyFunc(key)
}

// DeleteContractByKeyCalls gets all the calls that were made to DeleteContractByKey.
// Check the length with:
//
//	len(mockedSession.DeleteContractByKeyCalls())
func (mock *SessionMock) DeleteContractByKeyCalls() []struct {
	Key string
} {
	var calls []struct {
		Key string
	}
	mock.lockDeleteContractByKey.RLock()
	calls = mock.calls.DeleteContractByKey
	mock.lockDeleteContractByKey.RUnlock()
	return calls
}

// DeleteStructByKey calls DeleteStructByKeyFunc.
func (mock *SessionMock) DeleteStructByKey(key string) error {
	if mock.DeleteStructByKeyFunc == nil {
		panic("SessionMock.DeleteStructByKeyFunc: method is nil but Session.DeleteStructByKey was just called")
	}
	callInfo := struct {
		Key string
	}{
		Key: key,
	}
	mock.lockDeleteStructByKey.Lock()
	mock.calls.DeleteStructByKey = append(mock.calls.DeleteStructByKey, callInfo)
	mock.lockDeleteStructByKey.Unlock()
	return mock.DeleteStructByKeyFunc(key)
}

// DeleteStructByKeyCalls gets all the calls that were made to DeleteStructByKey.
// Check the length with:
//
//	len(mockedSession.DeleteStructByKeyCalls())
func (mock *SessionMock) DeleteStructByKeyCalls() []struct {
	Key string
} {
	var calls []struct {
		Key string
	}
	mock.lockDeleteStructByKey.RLock()
	calls = mock.calls.DeleteStructByKey
	mock.lockDeleteStructByKey.RUnlock()
	return calls
}

// GetContractByKey calls GetContractByKeyFunc.
func (mock *SessionMock) GetContractByKey(key string) (*ast.ContractToken, error) {
	if mock.GetContractByKeyFunc == nil {
		panic("SessionMock.GetContractByKeyFunc: method is nil but Session.GetContractByKey was just called")
	}
	callInfo := struct {
		Key string
	}{
		Key: key,
	}
	mock.lockGetContractByKey.Lock()
	mock.calls.GetContractByKey = append(mock.calls.GetContractByKey, callInfo)
	mock.lockGetContractByKey.Unlock()
	return mock.GetContractByKeyFunc(key)
}

// GetContractByKeyCalls gets all the calls that were made to GetContractByKey.
// Check the length with:
//
//	len(mockedSession.GetContractByKeyCalls())
func (mock *SessionMock) GetContractByKeyCalls() []struct {
	Key string
} {
	var calls []struct {
		Key string
	}
	mock.lockGetContractByKey.RLock()
	calls = mock.calls.GetContractByKey
	mock.lockGetContractByKey.RUnlock()
	return calls
}

// GetStructByKey calls GetStructByKeyFunc.
func (mock *SessionMock) GetStructByKey(key string) ([]*ast.StructToken, error) {
	if mock.GetStructByKeyFunc == nil {
		panic("SessionMock.GetStructByKeyFunc: method is nil but Session.GetStructByKey was just called")
	}
	callInfo := struct {
		Key string
	}{
		Key: key,
	}
	mock.lockGetStructByKey.Lock()
	mock.calls.GetStructByKey = append(mock.calls.GetStructByKey, callInfo)
	mock.lockGetStructByKey.Unlock()
	return mock.GetStructByKeyFunc(key)
}

// GetStructByKeyCalls gets all the calls that were made to GetStructByKey.
// Check the length with:
//
//	len(mockedSession.GetStructByKeyCalls())
func (mock *SessionMock) GetStructByKeyCalls() []struct {
	Key string
} {
	var calls []struct {
		Key string
	}
	mock.lockGetStructByKey.RLock()
	calls = mock.calls.GetStructByKey
	mock.lockGetStructByKey.RUnlock()
	return calls
}

// ReleaseStructObjectReadLock calls ReleaseStructObjectReadLockFunc.
func (mock *SessionMock) ReleaseStructObjectReadLock(structName string, keys ...[]byte) error {
	if mock.ReleaseStructObjectReadLockFunc == nil {
		panic("SessionMock.ReleaseStructObjectReadLockFunc: method is nil but Session.ReleaseStructObjectReadLock was just called")
	}
	callInfo := struct {
		StructName string
		Keys       [][]byte
	}{
		StructName: structName,
		Keys:       keys,
	}
	mock.lockReleaseStructObjectReadLock.Lock()
	mock.calls.ReleaseStructObjectReadLock = append(mock.calls.ReleaseStructObjectReadLock, callInfo)
	mock.lockReleaseStructObjectReadLock.Unlock()
	return mock.ReleaseStructObjectReadLockFunc(structName, keys...)
}

// ReleaseStructObjectReadLockCalls gets all the calls that were made to ReleaseStructObjectReadLock.
// Check the length with:
//
//	len(mockedSession.ReleaseStructObjectReadLockCalls())
func (mock *SessionMock) ReleaseStructObjectReadLockCalls() []struct {
	StructName string
	Keys       [][]byte
} {
	var calls []struct {
		StructName string
		Keys       [][]byte
	}
	mock.lockReleaseStructObjectReadLock.RLock()
	calls = mock.calls.ReleaseStructObjectReadLock
	mock.lockReleaseStructObjectReadLock.RUnlock()
	return calls
}

// ReleaseStructObjectWriteLock calls ReleaseStructObjectWriteLockFunc.
func (mock *SessionMock) ReleaseStructObjectWriteLock(structName string, keys ...[]byte) error {
	if mock.ReleaseStructObjectWriteLockFunc == nil {
		panic("SessionMock.ReleaseStructObjectWriteLockFunc: method is nil but Session.ReleaseStructObjectWriteLock was just called")
	}
	callInfo := struct {
		StructName string
		Keys       [][]byte
	}{
		StructName: structName,
		Keys:       keys,
	}
	mock.lockReleaseStructObjectWriteLock.Lock()
	mock.calls.ReleaseStructObjectWriteLock = append(mock.calls.ReleaseStructObjectWriteLock, callInfo)
	mock.lockReleaseStructObjectWriteLock.Unlock()
	return mock.ReleaseStructObjectWriteLockFunc(structName, keys...)
}

// ReleaseStructObjectWriteLockCalls gets all the calls that were made to ReleaseStructObjectWriteLock.
// Check the length with:
//
//	len(mockedSession.ReleaseStructObjectWriteLockCalls())
func (mock *SessionMock) ReleaseStructObjectWriteLockCalls() []struct {
	StructName string
	Keys       [][]byte
} {
	var calls []struct {
		StructName string
		Keys       [][]byte
	}
	mock.lockReleaseStructObjectWriteLock.RLock()
	calls = mock.calls.ReleaseStructObjectWriteLock
	mock.lockReleaseStructObjectWriteLock.RUnlock()
	return calls
}

// Rollback calls RollbackFunc.
func (mock *SessionMock) Rollback() error {
	if mock.RollbackFunc == nil {
		panic("SessionMock.RollbackFunc: method is nil but Session.Rollback was just called")
	}
	callInfo := struct {
	}{}
	mock.lockRollback.Lock()
	mock.calls.Rollback = append(mock.calls.Rollback, callInfo)
	mock.lockRollback.Unlock()
	return mock.RollbackFunc()
}

// RollbackCalls gets all the calls that were made to Rollback.
// Check the length with:
//
//	len(mockedSession.RollbackCalls())
func (mock *SessionMock) RollbackCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockRollback.RLock()
	calls = mock.calls.Rollback
	mock.lockRollback.RUnlock()
	return calls
}

// StructTombstones calls StructTombstonesFunc.
func (mock *SessionMock) StructTombstones() (map[string]string, []*ast.StructToken, error) {
	if mock.StructTombstonesFunc == nil {
		panic("SessionMock.StructTombstonesFunc: method is nil but Session.StructTombstones was just called")
	}
	callInfo := struct {
	}{}
	mock.lockStructTombstones.Lock()
	mock.calls.StructTombstones = append(mock.calls.StructTombstones, callInfo)
	mock.lockStructTombstones.Unlock()
	return mock.StructTombstonesFunc()
}

// StructTombstonesCalls gets all the calls that were made to StructTombstones.
// Check the length with:
//
//	len(mockedSession.StructTombstonesCalls())
func (mock *SessionMock) StructTombstonesCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockStructTombstones.RLock()
	calls = mock.calls.StructTombstones
	mock.lockStructTombstones.RUnlock()
	return calls
}

// Structs calls StructsFunc.
func (mock *SessionMock) Structs() ([]*ast.StructToken, error) {
	if mock.StructsFunc == nil {
		panic("SessionMock.StructsFunc: method is nil but Session.Structs was just called")
	}
	callInfo := struct {
	}{}
	mock.lockStructs.Lock()
	mock.calls.Structs = append(mock.calls.Structs, callInfo)
	mock.lockStructs.Unlock()
	return mock.StructsFunc()
}

// StructsCalls gets all the calls that were made to Structs.
// Check the length with:
//
//	len(mockedSession.StructsCalls())
func (mock *SessionMock) StructsCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockStructs.RLock()
	calls = mock.calls.Structs
	mock.lockStructs.RUnlock()
	return calls
}

// WriteContract calls WriteContractFunc.
func (mock *SessionMock) WriteContract(contract *ast.ContractToken) error {
	if mock.WriteContractFunc == nil {
		panic("SessionMock.WriteContractFunc: method is nil but Session.WriteContract was just called")
	}
	callInfo := struct {
		Contract *ast.ContractToken
	}{
		Contract: contract,
	}
	mock.lockWriteContract.Lock()
	mock.calls.WriteContract = append(mock.calls.WriteContract, callInfo)
	mock.lockWriteContract.Unlock()
	return mock.WriteContractFunc(contract)
}

// WriteContractCalls gets all the calls that were made to WriteContract.
// Check the length with:
//
//	len(mockedSession.WriteContractCalls())
func (mock *SessionMock) WriteContractCalls() []struct {
	Contract *ast.ContractToken
} {
	var calls []struct {
		Contract *ast.ContractToken
	}
	mock.lockWriteContract.RLock()
	calls = mock.calls.WriteContract
	mock.lockWriteContract.RUnlock()
	return calls
}
