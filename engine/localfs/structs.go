// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package localfs

import (
	"encoding/binary"
	"os"
	"path/filepath"

	"remixdb.io/engine/ifaces/planner"
)

func parseStructFields(b []byte, name string, inDB bool) (planner.Struct, []byte) {
	// Make sure the length is correct.
	if 4 > len(b) {
		// Panic since this is a corrupt database.
		panic("corrupt database - unable to read structs file")
	}
	fieldsCount := binary.LittleEndian.Uint32(b[:4])

	// Slice off the first 4 bytes.
	b = b[4:]

	// Make a slice of fields.
	fields := make([]*planner.StructField, fieldsCount)

	// Loop through the fields.
	for i := uint32(0); i < fieldsCount; i++ {
		// Get the field name length.
		if len(b) < 2 {
			// Panic since this is a corrupt database.
			panic("corrupt database - unable to read structs file")
		}

		// Get the name length.
		nameLen := binary.LittleEndian.Uint16(b[:2])

		// Slice off the first 2 bytes.
		b = b[2:]

		// Get the name.
		if len(b) < int(nameLen) {
			// Panic since this is a corrupt database.
			panic("corrupt database - unable to read structs file")
		}

		// Get the name.
		fieldName := string(b[:nameLen])

		// Slice off the name.
		b = b[nameLen:]

		// Get the type length.
		if len(b) < 2 {
			// Panic since this is a corrupt database.
			panic("corrupt database - unable to read structs file")
		}

		// Get the type length.
		typeLen := binary.LittleEndian.Uint16(b[:2])

		// Slice off the first 2 bytes.
		b = b[2:]

		// Get the type.
		if len(b) < int(typeLen) {
			// Panic since this is a corrupt database.
			panic("corrupt database - unable to read structs file")
		}

		// Get the type.
		fieldType := string(b[:typeLen])

		// Slice off the type.
		b = b[typeLen:]

		//
	}
}

func (ph *planHandler) Structs() []planner.Struct {
	// Get it from the transaction if we have it.
	if ph.structs != nil {
		return ph.structs
	}

	// Get it from the database.
	b, err := ph.e.m.ReadFile("structs", "root", 0, func() ([]byte, error) {
		fp := filepath.Join(ph.e.fp, "structs")
		return os.ReadFile(fp)
	})
	if err != nil {
		if os.IsNotExist(err) {
			// No structs.
			return []planner.Struct{}
		}

		// FS error. We should panic.
		panic(err)
	}

	// Make sure the length is correct.
	if 4 > len(b) {
		// Panic since this is a corrupt database.
		panic("corrupt database - unable to read structs file")
	}
	structsCount := binary.LittleEndian.Uint32(b[:4])

	// Slice off the first 4 bytes.
	b = b[4:]

	// Make a slice of structs.
	structs := make([]planner.Struct, structsCount)

	// Loop through the structs.
	for i := uint32(0); i < structsCount; i++ {
		// Get if this is in DB.
		if len(b) < 3 {
			// Panic since this is a corrupt database.
			panic("corrupt database - unable to read structs file")
		}
		inDB := b[0] == 1

		// Slice off the first byte.
		b = b[1:]

		// Get the name length.
		if len(b) < 2 {
			// Panic since this is a corrupt database.
			panic("corrupt database - unable to read structs file")
		}
		nameLen := binary.LittleEndian.Uint16(b[:2])

		// Slice off the first 2 bytes.
		b = b[2:]

		// Get the name.
		if len(b) < int(nameLen) {
			// Panic since this is a corrupt database.
			panic("corrupt database - unable to read structs file")
		}
		name := string(b[:nameLen])

		// Slice off the name.
		b = b[nameLen:]

		// Parse the struct fields in the file.
		structs[i], b = parseStructFields(b, name, inDB)
	}

	// Return the structs.
	return structs
}

func (ph *planHandler) Struct(name string) planner.Struct {
	for _, st := range ph.Structs() {
		if st.Name() == name {
			return st
		}
	}
	return nil
}
