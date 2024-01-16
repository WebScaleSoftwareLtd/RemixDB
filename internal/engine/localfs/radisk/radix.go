// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package radisk

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// TreeIO is used to define the interface for the tree IO.
type TreeIO struct {
	FS Filesystem
}

type branch struct {
	// Points to either 0x00 (no data), or a bytes location with a uint32 little endian length at the start.
	// If the length is 0, it should read the next 4 bytes as uint32 little endian and load that page.
	valuePtr uint32

	// Comprised of the key length (uint32 little endian) and the key itself.
	key []byte

	// Defines the children page and index where the child branches are stored. If both are zero, there
	// are no children. At the index should be a uint32 little endian length of branches, followed by
	// that many branches.
	childrenPage  uint32
	childrenIndex uint32
}

// ErrInvalidData is used to define the error when the data is invalid.
var ErrInvalidData = errors.New("invalid data")

func readBranch(data []byte) (b branch, n int, err error) {
	if len(data) < 8 {
		err = ErrInvalidData
		return
	}
	b.valuePtr = binary.LittleEndian.Uint32(data)

	keyLen := binary.LittleEndian.Uint32(data[4:])
	n = 8 + int(keyLen)
	data = data[8:]
	if len(data) < int(keyLen) {
		err = ErrInvalidData
		return
	}
	b.key = data[:keyLen]
	data = data[keyLen:]

	if len(data) < 8 {
		err = ErrInvalidData
		return
	}
	b.childrenPage = binary.LittleEndian.Uint32(data)
	b.childrenIndex = binary.LittleEndian.Uint32(data[4:])
	n += 8
	return
}

func getBytesState(possiblePrefix, key []byte) int {
	state := 0
	switch {
	case string(key) == string(possiblePrefix):
		state = 1
	case bytes.HasPrefix(key, possiblePrefix):
		state = 2
	}
	return state
}

type stack[T any] struct {
	prev *stack[T]
	data T
}

type branches struct {
	offset uint32
	count  uint32
}

// GetValue is used to get the value from the tree. Returns ErrNotFound if the value does not exist.
func (t TreeIO) GetValue(key []byte) ([]byte, error) {
	// Get the root page.
	page, err := t.FS.GetPage(0)
	if err != nil {
		return nil, err
	}

	// Defines items used in the loop.
	var pageNum uint32
	offsetStack := &stack[branches]{data: branches{offset: 0, count: 1}}

	// Start the loop.
	for offsetStack != nil {
		// Pop one off the stack.
		s := *offsetStack
		offsetStack = s.prev
		count := s.data.count
		offset := s.data.offset

		for i := uint32(0); i < count; i++ {
			// Read the branch.
			b, n, err := readBranch(page[offset:])
			if err != nil {
				return nil, err
			}
			offset += uint32(n)

			// Get our state.
			state := getBytesState(b.key, key)
			if state == 0 {
				// If we are in state 0, we need to go to the next branch.
				continue
			}

			// If we are in state 1, we have found the value.
			if state == 1 {
				if b.valuePtr == 0 {
					// If valuePtr is 0, there is no value.
					return nil, ErrNotFound
				}

				// Jump to where it is and try to read 4 bytes.
				if len(page) < int(b.valuePtr+4) {
					return nil, ErrInvalidData
				}
				valueLen := binary.LittleEndian.Uint32(page[b.valuePtr:])
				if valueLen == 0 {
					// Read the next 4 bytes as the page number.
					if len(page) < int(b.valuePtr+8) {
						return nil, ErrInvalidData
					}

					// Load the page and return it.
					return t.FS.GetPage(uint64(binary.LittleEndian.Uint32(page[b.valuePtr+4:])))
				}

				// Read the value.
				if len(page) < int(b.valuePtr+4+valueLen) {
					return nil, ErrInvalidData
				}
				return page[b.valuePtr+4 : b.valuePtr+4+valueLen], nil
			}

			// Chop off the prefix.
			key = key[len(b.key):]

			// Go to the children.
			if b.childrenPage != pageNum {
				// Load the page.
				page, err = t.FS.GetPage(uint64(b.childrenPage))
				if err != nil {
					return nil, err
				}
				pageNum = b.childrenPage
			}

			// Get the uint32 LE at the index.
			if len(page) < int(b.childrenIndex+4) {
				return nil, ErrInvalidData
			}
			childrenLen := binary.LittleEndian.Uint32(page[b.childrenIndex:])
			offset = b.childrenIndex + 4

			// Push the current offset onto the stack.
			offsetStack = &stack[branches]{prev: offsetStack, data: branches{offset: offset, count: childrenLen}}
		}
	}

	// Not found.
	return nil, ErrNotFound
}
