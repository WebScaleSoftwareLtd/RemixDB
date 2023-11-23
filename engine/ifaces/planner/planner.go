// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package planner

import "errors"

// StructForeignKey is used to define a foreign key for a struct field.
type StructForeignKey struct {
	// Struct is the struct that the foreign key is referencing.
	Struct Struct

	// Field is the field that the foreign key is referencing.
	Field string

	// OnDelete is the action to take when the foreign key is deleted.
	OnDelete string

	// OnUpdate is the action to take when the foreign key is updated.
	OnUpdate string
}

// StructField is used to define a field for a struct.
type StructField struct {
	// Name is the name of the field.
	Name string

	// Type is the type of the field.
	Type string

	// Index defines if this field is indexed.
	Index bool

	// ForeignKey defines if this field is using a foreign key.
	ForeignKey *StructForeignKey

	// Optional defines if this field is optional.
	Optional bool

	// BelongsTo is the struct that the field belongs to.
	BelongsTo Struct

	// FieldsReliant is the fields reliant on this field.
	FieldsReliant []*StructField

	// MappingsReliant is the mappings reliant on this field.
	MappingsReliant []Mapping

	// ContractsReliant is the contracts reliant on this field.
	ContractsReliant []Contract
}

// FieldTombstone is used to define a tombstone for a field.
type FieldTombstone struct {
	// Struct is the struct that the field belongs to.
	Struct Struct

	// Field is the field that is a tombstone.
	Field string

	// Type is the type of the field.
	Type string
}

// Struct is used to define a struct for the planner.
type Struct interface {
	// Name is used to get the name of the struct.
	Name() string

	// InDB is used to define if the struct is in the database.
	InDB() bool

	// Fields is used to get the fields of the struct.
	Fields() []*StructField

	// FieldTombstones is used to get the field tombstones of the struct.
	FieldTombstones() []*FieldTombstone

	// FieldsReliant is used to get the fields reliant on this struct.
	FieldsReliant() []*StructField

	// MappingsReliant is used to get the mappings reliant on this struct.
	MappingsReliant() []Mapping

	// ContractsReliant is used to get the contracts reliant on this struct.
	ContractsReliant() []Contract

	// Drop is used to drop the struct from the database. Returns
	// an error if anything is reliant on this struct.
	Drop() error

	// AddField is used to add a field to the struct. Returns an error
	// if the field already exists or if the field is named after a
	// tombstone with the wrong type.
	AddField(field *StructField) error

	// RemoveField is used to remove a field from the struct. Returns
	// an error if the field does not exist or if anything is reliant
	// on the field.
	RemoveField(fieldName string) error

	// RenameField is used to rename a field in the struct. Returns
	// an error if the field does not exist or if the new field name
	// already exists. The old field name will become a tombstone.
	RenameField(fieldName, newFieldName string) error

	// IndexField is used to index a field in the struct. Returns
	// an error if the field does not exist or if the field is already
	// indexed.
	IndexField(fieldName string) error

	// UnindexField is used to unindex a field in the struct. Returns
	// an error if the field does not exist or if the field is not
	// indexed.
	UnindexField(fieldName string) error

	// AddFieldForeignKey is used to add a foreign key to a field in
	// the struct. Returns an error if the field does not exist or if
	// the field already has a foreign key.
	AddFieldForeignKey(foreignKey *StructForeignKey) error

	// RemoveFieldForeignKey is used to remove a foreign key from a
	// field in the struct. Returns an error if the field does not
	// exist or if the field does not have a foreign key.
	RemoveFieldForeignKey(fieldName string) error

	// MakeFieldOptional is used to make a field optional in the struct.
	// Returns an error if the field does not exist, if the field is
	// already optional, or if it is relied upon by a mapping or contract.
	MakeFieldOptional(fieldName string) error

	// MakeFieldRequired is used to make a field required in the struct.
	// Returns an error if the field does not exist, if the field is
	// already required, or if it is relied upon by a mapping or contract.
	MakeFieldRequired(fieldName string) error
}

// MappingInner is used to define an inner mapping for a mapping.
type MappingInner struct {
	// Mapping is the mapping that the mapping is referencing.
	Mapping Mapping

	// Field is the field that the mapping is referencing.
	Key string

	// Value is the value that the mapping is referencing.
	// It is either MappingInner or string.
	Value any
}

// ErrMappingWrongType is used when a mapping value is of the wrong type.
var ErrMappingWrongType = errors.New("mapping value is of the wrong type compared to before")

// ErrMappingValueDoesNotExist is used when a mapping value does not exist.
var ErrMappingValueDoesNotExist = errors.New("mapping value does not exist")

// Mapping is used to define a mapping for the planner.
type Mapping interface {
	// Name is used to get the name of the mapping.
	Name() string

	// Key is used to get the key of the mapping.
	Key() string

	// Value is used to get the value of the mapping.
	// It is either MappingInner or string.
	Value() any

	// EditKey is used to edit the key of the mapping. Returns
	// an error if the key already exists.
	EditKey(key string) error

	// EditValue is used to edit the value of the mapping. Returns
	// an error if the value is of the wrong type compared to before.
	// Expected type is either MappingInner or string.
	EditValue(value any) error

	// Drop is used to drop the mapping from the database.
	Drop() error
}

// ContractArgument is used to define an argument for a contract.
type ContractArgument struct {
	// Name is the name of the argument.
	Name string

	// Type is the type of the argument.
	Type string
}

// Contract is used to define a contract for the planner.
type Contract interface {
	// Name is used to get the name of the contract.
	Name() string

	// Argument is used to get the argument of the contract. Returns
	// nil if the contract does not have an argument.
	Argument() *ContractArgument

	// ChangeArgumentName is used to change the name of the argument.
	// Returns an error if the argument does not exist or if the new
	// name already exists.
	ChangeArgumentName(newName string) error

	// Throws is used to get the exceptions that the contract can raise.
	Throws() []string

	// ReplaceContract is used to replace the contract with another
	// contract. Returns an error if the contract does not exist.
	ReplaceContract(c Contract) error

	// Instructions is used to get the instructions of the contract.
	// Returns nil if the contract does not exist. See instructions.go
	// for all of the instructions.
	Instructions() ([]any, error)

	// Drop is used to drop the contract from the database.
	Drop() error
}

// ErrPlannerInReadLock is used when a planner is in a read lock and cannot
// stage a plan.
var ErrPlannerInReadLock = errors.New("planner is in a read lock")

// PlanHandler is used to define the handler when we have a lock for a planner.
type PlanHandler interface {
	// Unlock is used to unlock the planner. This should be called when the
	// planner is no longer needed.
	Unlock()

	// Struct is used to get a struct from the planner. This will return
	// a nil pointer if the struct does not exist.
	Struct(name string) Struct

	// Structs is used to get all structs from the planner.
	Structs() []Struct

	// StructViolatesTombstone is used to check if a struct violates a tombstone.
	StructViolatesTombstone(s Struct) bool

	// AddStruct is used to add a struct to the planner. This will return
	// an error if the struct already exists.
	AddStruct(s Struct) error

	// Mapping is used to get a mapping from the planner. This will return
	// a nil pointer if the mapping does not exist.
	Mapping(name string) Mapping

	// Mappings is used to get all mappings from the planner.
	Mappings() []Mapping

	// AddMapping is used to add a mapping to the planner. This will return
	// an error if the mapping already exists.
	AddMapping(m Mapping) error

	// MappingViolatesTombstone is used to check if a mapping violates a tombstone.
	MappingViolatesTombstone(m Mapping) bool

	// Contract is used to get a contract from the planner. This will return
	// a nil pointer if the contract does not exist.
	Contract(name string) Contract

	// ContractViolatesTombstone is used to check if a contract violates a tombstone.
	ContractViolatesTombstone(c Contract) bool

	// Contracts is used to get all contracts from the planner.
	Contracts() []Contract

	// AddContract is used to add a contract to the planner. This will return
	// an error if the contract already exists.
	AddContract(c Contract) error

	// StagePlan is used to stage a plan for the planner to the database. This
	// will return an ErrPlannerInReadLock if the planner is in a read lock.
	StagePlan() error
}

// Planner is used to define the structure of a planner for database tasks.
type Planner interface {
	// AcquirePlannerLock is used to acquire a write lock for the planner.
	AcquirePlannerLock() PlanHandler

	// AcquirePlannerReadLock is used to acquire a read lock for the planner.
	// A read lock cannot stage a plan, but can read the current schema.
	AcquirePlannerReadLock() PlanHandler
}
