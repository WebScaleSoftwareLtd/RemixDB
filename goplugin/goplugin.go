//go:build !windows

// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package goplugin

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"runtime"
	"strings"

	"remixdb.io/goplugin/cgocheck"
	_ "remixdb.io/internal/httpproxypatch"
	"remixdb.io/logger"
)

// GoPluginCompiler is used to define the Go plugin compiler. This turns the specified
// Go code into a plugin that can be used by RemixDB.
type GoPluginCompiler struct {
	logger     logger.Logger
	path       string
	projectZip []byte
}

// ExecutionError is used to define an error that occurred during execution.
type ExecutionError struct {
	exitCode int
	data     []byte
}

// Error is used to return the error as a string.
func (e ExecutionError) Error() string {
	return fmt.Sprintf("execution error: status %d: %s", e.exitCode, string(e.data))
}

// Compile is used to compile the Go plugin or return a cached version. It is compiled
// within the project zip specified. This is thread safe.
func (g GoPluginCompiler) Compile(code string) (*plugin.Plugin, error) {
	// Get the filename of the plugin.
	shaB := sha256.Sum256([]byte(code))
	pluginName := base64.URLEncoding.EncodeToString(shaB[:])

	// Load the plugin if it exists.
	pluginBinPath := filepath.Join(g.path, "plugins", pluginName+".so")
	if _, err := os.Stat(pluginBinPath); err == nil {
		return plugin.Open(pluginBinPath)
	}

	// Create a temporary directory.
	tempDir, err := os.MkdirTemp("", "remixdb-goplugin-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	// Extract the project zip into the temporary directory.
	r, err := zip.NewReader(bytes.NewReader(g.projectZip), int64(len(g.projectZip)))
	if err != nil {
		return nil, err
	}
	if err := handleZipReader(tempDir, r); err != nil {
		return nil, err
	}

	// Make a plugingen directory.
	pluginGenDir := filepath.Join(tempDir, "plugingen")
	if err := os.MkdirAll(pluginGenDir, 0755); err != nil {
		return nil, err
	}

	// Write the code to the file.
	goFile := filepath.Join(pluginGenDir, pluginName+".go")
	if err := os.WriteFile(goFile, []byte(code), 0644); err != nil {
		return nil, err
	}

	// Define the arguments for the compiler.
	args := []string{
		"build", "-buildmode=plugin", "-o", pluginBinPath, goFile,
	}
	envStrings := os.Environ()
	env := map[string]string{}
	for _, v := range envStrings {
		// Remove any variables that start with "GO".
		if strings.HasPrefix(v, "GO") {
			continue
		}

		split := strings.SplitN(v, "=", 2)
		env[split[0]] = split[1]
	}
	env["GOOS"] = runtime.GOOS
	env["GOARCH"] = runtime.GOARCH
	env["GOMODCACHE"] = filepath.Join(g.path, "cache")
	env["GOPATH"] = filepath.Join(g.path, "go")
	cgoEnabled := "0"
	if cgocheck.CGO {
		cgoEnabled = "1"
	}
	env["CGO_ENABLED"] = cgoEnabled
	bin := filepath.Join(g.path, "go", "bin", "go")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	envStrings = make([]string, len(env))
	i := 0
	for k, v := range env {
		envStrings[i] = k + "=" + v
		i++
	}

	// Run the compiler.
	cmd := exec.Command(bin, args...)
	cmd.Dir = tempDir
	cmd.Env = envStrings
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, ExecutionError{
				exitCode: exitErr.ExitCode(),
				data:     buf.Bytes(),
			}
		}
		return nil, err
	}

	// Load the plugin.
	return plugin.Open(pluginBinPath)
}

func defaultPath() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		// The home directory is required for RemixDB to work if no env is specified.
		panic(err)
	}

	fp := filepath.Join(homedir, ".remixdb", "goplugin")
	if err := os.MkdirAll(fp, 0755); err != nil {
		// Must be able to create the directory.
		panic(err)
	}

	return fp
}

func handleZipReader(path string, zipReader *zip.Reader) error {
	// Extract each file from the zip archive
	for _, file := range zipReader.File {
		filePath := path + "/" + file.Name

		if file.FileInfo().IsDir() {
			// Create directories
			err := os.MkdirAll(filePath, file.Mode())
			if err != nil {
				return err
			}
			continue
		}

		// Create the file
		newFile, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer newFile.Close()

		// Open the file in the zip archive
		zipFile, err := file.Open()
		if err != nil {
			return err
		}
		defer zipFile.Close()

		// Copy the contents of the file from the zip archive to the new file
		_, err = io.Copy(newFile, zipFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func extractZipInto(f *os.File, path string) error {
	_, _ = f.Seek(0, io.SeekStart)
	stat, err := f.Stat()
	if err != nil {
		return err
	}

	zipReader, err := zip.NewReader(f, stat.Size())
	if err != nil {
		return err
	}

	return handleZipReader(path, zipReader)
}

func extractTarGzInto(f *os.File, path string) error {
	_, err := f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	// Open the gzip reader
	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	// Open the tar reader
	tarReader := tar.NewReader(gzipReader)

	// Extract each file from the tar archive
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break // end of tar archive
		}

		if err != nil {
			return err
		}

		// Calculate the full file path
		filePath := filepath.Join(path, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directories
			err := os.MkdirAll(filePath, header.FileInfo().Mode())
			if err != nil {
				return err
			}
		case tar.TypeReg:
			// Create the file
			newFile, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, header.FileInfo().Mode())
			if err != nil {
				return err
			}
			defer newFile.Close()

			// Copy the contents of the file from the tar archive to the new file
			_, err = io.Copy(newFile, tarReader)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func downloadGo(path, goVersionFileExpected string, logger logger.Logger) {
	ch := make(chan string, 1)
	logger.Info("Downloading "+goVersionFileExpected, ch)

	// Make sure <path>/go doesn't exist.
	if err := os.RemoveAll(filepath.Join(path, "go")); err != nil {
		if !os.IsNotExist(err) {
			// Must be able to remove the directory.
			panic(err)
		}
	}

	// Generate the Go download URL.
	url := "https://golang.org/dl/" + runtime.Version() + "." + runtime.GOOS + "-" + runtime.GOARCH + "."
	if runtime.GOOS == "windows" {
		url += "zip"
	} else {
		url += "tar.gz"
	}

	// Download the Go archive.
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		panic("invalid status code for Go download: " + resp.Status)
	}
	f, err := os.CreateTemp("", "remixdb-go-download-")
	if err != nil {
		panic(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		panic(err)
	}

	// Extract the Go archive.
	if runtime.GOOS == "windows" {
		if err := extractZipInto(f, path); err != nil {
			panic(err)
		}
	} else {
		if err := extractTarGzInto(f, path); err != nil {
			panic(err)
		}
	}

	// Log that we are done!
	ch <- "done!"
}

// Extracts the cache.
func extractCache(path string, cacheZip []byte) {
	// Clean <path>/cache.
	if err := os.RemoveAll(filepath.Join(path, "cache")); err != nil {
		if !os.IsNotExist(err) {
			// Must be able to remove the directory.
			panic(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(path, "cache"), 0755); err != nil {
		// Must be able to create the directory.
		panic(err)
	}

	// Extract the cache.
	r, err := zip.NewReader(bytes.NewReader(cacheZip), int64(len(cacheZip)))
	if err != nil {
		panic(err)
	}
	if err := handleZipReader(filepath.Join(path, "cache"), r); err != nil {
		panic(err)
	}
}

// NewGoPluginCompiler is used to create a new Go plugin compiler. If path is empty, then it will try and use the
// environment or ~/.remixdb/goplugin. No other argument can be empty.
func NewGoPluginCompiler(logger logger.Logger, path string, cacheZip, projectZip []byte) GoPluginCompiler {
	// Add the label to the logger.
	logger = logger.Tag("goplugin")

	// Get the path for everything.
	if path == "" {
		path = os.Getenv("REMIXDB_GOPLUGIN_PATH")
		if path == "" {
			// Use the default path.
			path = defaultPath()
		} else {
			// Make sure the path is valid.
			if err := os.MkdirAll(path, 0755); err != nil {
				// Must be able to create the directory.
				panic(err)
			}
		}
	}

	// Check if the Go version changed since the last time we ran.
	b, _ := os.ReadFile(filepath.Join(path, ".go_version"))
	if b == nil {
		b = []byte{}
	}
	goVersionFileExpected := runtime.Version() + " / " + runtime.GOOS + " / " + runtime.GOARCH
	cleanPlugins := false
	if string(b) == goVersionFileExpected {
		// Log that we have done this already.
		logger.Info("Downloading "+goVersionFileExpected+"... cached!", nil)
	} else {
		// Download Go.
		downloadGo(path, goVersionFileExpected, logger)

		// Write the Go version file.
		if err := os.WriteFile(filepath.Join(path, ".go_version"), []byte(goVersionFileExpected), 0644); err != nil {
			panic(err)
		}

		// Set clean plugins to true.
		cleanPlugins = true
	}

	replaceCache := false
	var hashS string
	if !cleanPlugins {
		// Hash the cache zip.
		hasher := sha256.New()
		_, _ = hasher.Write(cacheZip)
		hash := hasher.Sum(nil)
		hashS = hex.EncodeToString(hash)

		// Check if the hash changed since the last time we ran.
		b, _ = os.ReadFile(filepath.Join(path, ".hash"))
		if b == nil {
			b = []byte{}
		}
		if string(b) == hashS {
			// Since they are the same, we don't need to clean the plugins.
			// Return the compiler.
			return GoPluginCompiler{
				logger:     logger,
				path:       path,
				projectZip: projectZip,
			}
		} else {
			// The cache changed, so we need to replace it. We will also clean the plugins.
			replaceCache = true
		}
	}

	// Clean the plugins.
	if err := os.RemoveAll(filepath.Join(path, "plugins")); err != nil {
		if !os.IsNotExist(err) {
			// Must be able to remove the directory.
			panic(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(path, "plugins"), 0755); err != nil {
		// Must be able to create the directory.
		panic(err)
	}

	// Return here if we don't need to replace the cache.
	if !replaceCache {
		return GoPluginCompiler{
			logger:     logger,
			path:       path,
			projectZip: projectZip,
		}
	}

	// Extract the cache.
	extractCache(path, cacheZip)

	// Write the hash file.
	if err := os.WriteFile(filepath.Join(path, ".hash"), []byte(hashS), 0644); err != nil {
		panic(err)
	}

	// Return the compiler.
	return GoPluginCompiler{
		logger:     logger,
		path:       path,
		projectZip: projectZip,
	}
}
