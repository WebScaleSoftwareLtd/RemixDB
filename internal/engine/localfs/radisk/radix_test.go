// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package radisk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockFSReader struct {
	pages map[uint64][]byte
}

func (m mockFSReader) GetPage(page uint64) ([]byte, error) {
	if data, ok := m.pages[page]; ok {
		return data, nil
	}
	return nil, ErrNotFound
}

func (m mockFSReader) SetPage(page uint64, data []byte) error {
	panic("read only")
}

var multiPageKey = map[uint64][]byte{
	0: {
		// No value on this branch
		0x00, 0x00, 0x00, 0x00,
		// Empty key
		0x00, 0x00, 0x00, 0x00,
		// Children page
		0x01, 0x00, 0x00, 0x00,
		// Children index
		0x01, 0x00, 0x00, 0x00,
	},
	1: {
		// Null byte at beginning of page for testing
		0x00,

		// Number of branches
		0x02, 0x00, 0x00, 0x00,

		// First branch
		0x00, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x00, 0x00,
		'b',
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,

		// Second branch
		0x01, 0x00, 0x00, 0x00, // This will error if we are reading the wrong branch
		0x03, 0x00, 0x00, 0x00, // Length is 3
		'a', 'b', 'c', // Part of the key
		69, 0, 0, 0, // Page 69
		0x00, 0x00, 0x00, 0x00, // Index 0
	},
	69: {
		// Number of branches
		0x01, 0x00, 0x00, 0x00,

		// First branch
		0x16, 0x00, 0x00, 0x00, // Value pointer
		0x02, 0x00, 0x00, 0x00, // Key length
		':', 'd', // Key
		0x00, 0x00, 0x00, 0x00, // Children page
		0x00, 0x00, 0x00, 0x00, // Children index

		// Value
		0x04, 0x00, 0x00, 0x00, 't', 'e', 's', 't',
	},
}

func TestTreeIO_GetValue(t *testing.T) {
	tests := []struct {
		name string

		pages map[uint64][]byte
		key   string

		wantsErr error
		expected string
	}{
		{
			name:     "no pages",
			pages:    map[uint64][]byte{},
			wantsErr: ErrNotFound,
		},
		{
			name: "no data",
			pages: map[uint64][]byte{
				0: {
					// Value pointer (no data)
					0x00, 0x00, 0x00, 0x00,
					// Key length
					0x00, 0x00, 0x00, 0x00,
					// Children page
					0x00, 0x00, 0x00, 0x00,
					// Children index
					0x00, 0x00, 0x00, 0x00,
				},
			},
			wantsErr: ErrNotFound,
		},
		{
			name: "blank key",
			pages: map[uint64][]byte{
				0: {
					// Value pointer (point to after the branch)
					0x10, 0x00, 0x00, 0x00,
					// Key length
					0x00, 0x00, 0x00, 0x00,
					// Children page
					0x00, 0x00, 0x00, 0x00,
					// Children index
					0x00, 0x00, 0x00, 0x00,
					// The value that is pointed to
					0x04, 0x00, 0x00, 0x00, 't', 'e', 's', 't', 'i', 'n', 'g',

					// Random data
					'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
				},
			},
			expected: "test",
		},
		{
			name: "blank key with value on another page",
			pages: map[uint64][]byte{
				0: {
					// Value pointer (point to after the branch)
					0x10, 0x00, 0x00, 0x00,
					// Key length
					0x00, 0x00, 0x00, 0x00,
					// Children page
					0x00, 0x00, 0x00, 0x00,
					// Children index
					0x00, 0x00, 0x00, 0x00,

					// Repersents a page switch
					0x00, 0x00, 0x00, 0x00,
					// The page that the value is on
					0x02, 0x00, 0x00, 0x00,
				},
				2: []byte("hello world"),
			},
			expected: "hello world",
		},
		{
			name: "key amongst many branches",
			pages: map[uint64][]byte{
				0: {
					// No value on this branch
					0x00, 0x00, 0x00, 0x00,
					// Key length
					0x01, 0x00, 0x00, 0x00,
					// Key
					'a',
					// Children page
					0x00, 0x00, 0x00, 0x00,
					// Children index
					0x11, 0x00, 0x00, 0x00,

					// Length of children
					0x02, 0x00, 0x00, 0x00,
					// First child
					0x00, 0x00, 0x00, 0x00,
					0x01, 0x00, 0x00, 0x00,
					'a',
					0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00,
					// Second child
					0x3A, 0x00, 0x00, 0x00,
					0x04, 0x00, 0x00, 0x00,
					'b', 'c', ':', 'd',
					0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00,

					// Data
					0x04, 0x00, 0x00, 0x00, 't', 'e', 's', 't',
				},
			},
			key:      "abc:d",
			expected: "test",
		},
		{
			name:     "key amongst pages",
			pages:    multiPageKey,
			key:      "abc:d",
			expected: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			io := TreeIO{
				FS: mockFSReader{
					pages: tt.pages,
				},
			}
			var b []byte
			if tt.expected != "" {
				b = []byte(tt.expected)
			}
			got, err := io.GetValue([]byte(tt.key))
			if tt.wantsErr != nil {
				assert.Equal(t, tt.wantsErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, b, got)
			}
		})
	}
}

func BenchmarkTreeIO_GetValue(b *testing.B) {
	io := TreeIO{
		FS: mockFSReader{
			pages: multiPageKey,
		},
	}
	for i := 0; i < b.N; i++ {
		_, _ = io.GetValue([]byte("abc:d"))
	}
}
