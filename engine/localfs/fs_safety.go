// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package localfs

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Handles any failed transactions.
func handleFailedTransactions(fp string) {
	// Handle if any structs were stuck in a condom.
	dataPath := filepath.Join(fp, "data")
	dir, err := os.ReadDir(dataPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No data directory.
			return
		}
		panic(err)
	}
	for _, f := range dir {
		// Handle pending bulk filesystem operations.
		handlePendingBulkFilesystemOperations(
			filepath.Join(dataPath, f.Name()),
		)
	}
}

// Handles pending bulk filesystem operations.
func handlePendingBulkFilesystemOperations(fp string) {
	// Check if the 'pending' directory exists.
	pending := filepath.Join(fp, "pending")
	if _, err := os.Stat(pending); err != nil {
		if os.IsNotExist(err) {
			// No pending filesystem operations.
			return
		}

		// We should panic.
		panic(err)
	}

	// Check if 'C' exists.
	if _, err := os.Stat(filepath.Join(pending, "C")); err != nil {
		if os.IsNotExist(err) {
			// Never fully committed. Probably crashed mid operation
			// but pre-commit. We should remove the directory.
			if err := os.RemoveAll(pending); err != nil {
				panic(err)
			}
			return
		}

		// We should panic.
		panic(err)
	}

	// Read the 'A' file.
	b, err := os.ReadFile(filepath.Join(pending, "A"))
	if err != nil {
		if os.IsNotExist(err) {
			// No actions.
			if err := os.RemoveAll(pending); err != nil {
				panic(err)
			}
			return
		}
		panic(err)
	}

	// Split the bytes by new line.
	ops := bytes.Split(b, []byte("\n"))
	for _, opLine := range ops {
		// Continue if the line is empty.
		if len(opLine) == 0 {
			continue
		}

		// Get the operation.
		op := opLine[0]

		// Get the file path.
		fpPart := string(opLine[1:])

		// Switch on the operation.
		switch op {
		case 'a':
			// Removes all files in a directory.
			if err := os.RemoveAll(filepath.Join(fp, fpPart)); err != nil {
				// We do not care about do not exist errors since it might
				// have crashed right after this.
				if !os.IsNotExist(err) {
					panic(err)
				}
			}
		case 'R':
			// Remove the file.
			if err := os.Remove(filepath.Join(fp, fpPart)); err != nil {
				// We do not care about do not exist errors since it might
				// have crashed right after this.
				if !os.IsNotExist(err) {
					panic(err)
				}
			}
		case 'W':
			// Write the file by copying it over.
			originalFp := filepath.Join(pending, "W", fpPart)
			b, err := os.ReadFile(originalFp)
			if err != nil {
				panic(err)
			}
			folder, _ := filepath.Split(fpPart)
			if err := os.MkdirAll(filepath.Join(fp, folder), 0755); err != nil {
				panic(err)
			}
			if err := os.WriteFile(filepath.Join(fp, fpPart), b, 0644); err != nil {
				panic(err)
			}
		case 'r':
			// Rename the file.
			var s []string
			if err := json.Unmarshal(opLine[1:], &s); err != nil {
				panic(err)
			}
			old := filepath.Join(fp, s[0])
			new := filepath.Join(fp, s[1])
			if err := os.Rename(old, new); err != nil {
				// We do not care about do not exist errors since it might
				// have crashed right after this.
				if !os.IsNotExist(err) {
					panic(err)
				}
			}
		}
	}

	// Remove the pending directory.
	if err := os.RemoveAll(pending); err != nil {
		panic(err)
	}
}

type fsOperation struct {
	// Fp1 is the first file path.
	Fp1 string

	// Fp2 is the second file path. Used for some operations.
	Fp2 string

	// B is the bytes. Used for some operations.
	B []byte

	// All is used for some operations.
	All bool
}

type fsOperationHandler struct {
	fp string

	ops []fsOperation
}

// Makes a bulk filesystem operation wrapper.
func makeBulkFilesystemOperationHandler(fp string) *fsOperationHandler {
	// Make the pending directory.
	pending := filepath.Join(fp, "pending")
	if err := os.MkdirAll(pending, 0755); err != nil {
		panic(err)
	}

	// Make the 'W' directory.
	if err := os.MkdirAll(filepath.Join(pending, "W"), 0755); err != nil {
		panic(err)
	}

	// Return the operation handler.
	return &fsOperationHandler{
		fp: fp,
	}
}

// WriteFile writes a file.
func (f *fsOperationHandler) WriteFile(rel string, b []byte) {
	// Add the operation.
	f.ops = append(f.ops, fsOperation{
		Fp1: rel,
		B:   b,
	})
	f.createTransactionFile()

	// Write the file.
	wFolder := filepath.Join(f.fp, "pending", "W")
	folder, _ := filepath.Split(rel)
	if err := os.MkdirAll(filepath.Join(wFolder, folder), 0755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(wFolder, rel), b, 0644); err != nil {
		panic(err)
	}
}

// DeleteFile deletes a file.
func (f *fsOperationHandler) DeleteFile(rel string) {
	// Add the operation.
	f.ops = append(f.ops, fsOperation{
		Fp1: rel,
	})
	f.createTransactionFile()
}

// RemoveAll removes all files in a directory.
func (f *fsOperationHandler) RemoveAll(rel string) {
	// Add the operation.
	f.ops = append(f.ops, fsOperation{
		Fp1: rel,
		All: true,
	})
	f.createTransactionFile()
}

// Rename renames a file.
func (f *fsOperationHandler) Rename(old, new string) {
	f.ops = append(f.ops, fsOperation{
		Fp1: old,
		Fp2: new,
	})
	f.createTransactionFile()
}

// Creates a transaction file.
func (f *fsOperationHandler) createTransactionFile() {
	// Create the strings.
	s := make([]string, len(f.ops))
	for i, op := range f.ops {
		// Get the operation type.
		o := ""
		if op.All {
			o = "a"
		} else {
			if op.B == nil {
				// Either rename or delete.
				if op.Fp2 == "" {
					// Delete.
					o = "R"
				} else {
					// Rename.
					o = "r"
				}
			} else {
				// Write.
				o = "W"
			}
		}

		if o == "r" {
			// If it is rename, we need to make a slice of the file paths.
			b, _ := json.Marshal([]string{op.Fp1, op.Fp2})
			s[i] = o + string(b)
		} else {
			// Otherwise, we just need to add the file path.
			s[i] = o + op.Fp1
		}
	}

	// Write the file.
	if err := os.WriteFile(
		filepath.Join(f.fp, "pending", "A"),
		[]byte(strings.Join(s, "\n")), 0644,
	); err != nil {
		panic(err)
	}
}

// Rollback is used to rollback the transaction.
func (f *fsOperationHandler) Rollback() {
	// Remove the entire pending directory.
	if err := os.RemoveAll(filepath.Join(f.fp, "pending")); err != nil {
		panic(err)
	}
}

// Commit is used to commit the transaction.
func (f *fsOperationHandler) Commit() {
	// Create the 'C' file.
	if err := os.WriteFile(
		filepath.Join(f.fp, "pending", "C"),
		[]byte{}, 0644,
	); err != nil {
		panic(err)
	}

	// Go through the operations.
	for _, op := range f.ops {
		if op.All {
			// Remove all files in the directory.
			if err := os.RemoveAll(filepath.Join(f.fp, op.Fp1)); err != nil {
				// We do not care about do not exist errors since it might
				// have crashed right after this.
				if !os.IsNotExist(err) {
					panic(err)
				}
			}
		} else {
			if op.B == nil {
				// Either rename or delete.
				if op.Fp2 == "" {
					// Delete.
					if err := os.Remove(filepath.Join(f.fp, op.Fp1)); err != nil {
						// We do not care about do not exist errors since it might
						// have crashed right after this.
						if !os.IsNotExist(err) {
							panic(err)
						}
					}
				} else {
					// Rename.
					if err := os.Rename(
						filepath.Join(f.fp, op.Fp1),
						filepath.Join(f.fp, op.Fp2),
					); err != nil {
						// We do not care about do not exist errors since it might
						// have crashed right after this.
						if !os.IsNotExist(err) {
							panic(err)
						}
					}
				}
			} else {
				// Write.
				if err := os.WriteFile(filepath.Join(f.fp, op.Fp1), op.B, 0644); err != nil {
					panic(err)
				}
			}
		}
	}

	// Remove the entire pending directory.
	if err := os.RemoveAll(filepath.Join(f.fp, "pending")); err != nil {
		panic(err)
	}
}
