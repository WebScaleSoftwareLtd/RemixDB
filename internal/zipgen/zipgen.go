// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package zipgen

import (
	"archive/zip"
	"bytes"
	"fmt"
)

func processMap(prefix string, m map[string]any, w *zip.Writer) {
	for k, v := range m {
		switch x := v.(type) {
		case map[string]any:
			// Make the directory.
			h := &zip.FileHeader{
				Name:   prefix + k + "/",
				Method: zip.Store,
			}
			h.SetMode(0755)
			_, err := w.CreateHeader(h)
			if err != nil {
				panic(err)
			}

			// Process the map.
			processMap(prefix+k+"/", x, w)
		case []byte:
			// Make the file header.
			h := &zip.FileHeader{
				Name:   prefix + k,
				Method: zip.Store,
			}
			h.SetMode(0644)
			f, err := w.CreateHeader(h)
			if err != nil {
				panic(err)
			}

			// Write the file.
			_, err = f.Write(x)
			if err != nil {
				panic(err)
			}
		case string:
			// Make the file header.
			h := &zip.FileHeader{
				Name:   prefix + k,
				Method: zip.Store,
			}
			h.SetMode(0644)
			f, err := w.CreateHeader(h)
			if err != nil {
				panic(err)
			}

			// Write the file.
			_, err = f.Write([]byte(x))
			if err != nil {
				panic(err)
			}
		default:
			panic("unknown type: " + fmt.Sprint(x))
		}
	}
}

// CreateZip creates a zip file from a map. The files can be of type
// map[string]any (directory), or []byte/string (file).
func CreateZip(files map[string]any) []byte {
	buf := &bytes.Buffer{}
	w := zip.NewWriter(buf)
	defer w.Close()

	if files != nil {
		processMap("", files, w)
	}
	_ = w.Close()
	return buf.Bytes()
}
