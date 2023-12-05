// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"runtime"

	"remixdb.io/goplugin"
	"remixdb.io/logger"
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

func createZip(files map[string]any) []byte {
	buf := &bytes.Buffer{}
	w := zip.NewWriter(buf)
	defer w.Close()

	processMap("", files, w)
	_ = w.Close()
	return buf.Bytes()
}

func nonWindowsSetup(logger logger.Logger) {
	goplugin.NewGoPluginCompiler(logger, createZip(map[string]any{
		"lol": "hi",
	}), createZip(map[string]any{
		"go.mod": "module remixdb.io/x\n",
	}))
}

func main() {
	logger := logger.NewStdLogger()

	if runtime.GOOS != "windows" {
		nonWindowsSetup(logger)
	}
}
