// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package acid

import (
	"os"
	"path/filepath"
	"strconv"

	cp "github.com/otiai10/copy"
	"github.com/vmihailenco/msgpack/v5"
)

type actionType uint16

const (
	// actionTypeCreate is the action type for creating a file.
	// P1 is the relative path to the file. The relative path
	// can be added to 'd' inside the transaction folder to get
	// the contents of the file.
	actionTypeCreate actionType = iota

	// actionTypeDeleteAll is the action type for deleting all files.
	// P1 is the relative path to the folder or file.
	actionTypeDeleteAll

	// actionTypeDelete is the action type for deleting a file.
	// P1 is the relative path to the file.
	actionTypeDelete

	// actionTypeRename is the action type for renaming a file.
	// P1 is the relative path to the first file. P2 is the relative
	// path to the second file.
	actionTypeRename

	// actionTypeMkdir is the action type for creating a folder.
	// P1 is the relative path to the folder.
	actionTypeMkdir

	// actionTypeMkdirAll is the action type for creating a folder
	// and all its parents. P1 is the relative path to the folder.
	actionTypeMkdirAll
)

type journalAction struct {
	next *journalAction
	num  int

	// P1 is the path to the first file.
	P1 string

	// P2 is the path to the second file.
	P2 string

	// T is the type of action.
	T actionType
}

// Writes a journal action object to the queue and the disk.
func (t *Transaction) writeJournalAction(action journalAction) {
	// Add it to the queue.
	if t.journalActionsStart == nil {
		t.journalActionsStart = &action
		t.journalActionsEnd = &action
	} else {
		t.journalActionsEnd.next = &action
		t.journalActionsEnd = &action
	}

	// Write it to the disk.
	count := t.count
	t.count++
	action.num = count
	b, err := msgpack.Marshal(action)
	if err != nil {
		panic(err)
	}
	fp := filepath.Join(t.getTransactionFolder(), strconv.Itoa(count)+".J")
	if err := os.WriteFile(fp, b, 0644); err != nil {
		panic(err)
	}
}

// WriteFile is used to write a file to the data folder. The path should be relative to the data folder.
func (t *Transaction) WriteFile(path string, data []byte) {
	t.writeJournalAction(journalAction{
		T:  actionTypeCreate,
		P1: path,
	})
	fpJoin := filepath.Join(t.getTransactionFolder(), "d", path)
	if err := os.MkdirAll(filepath.Dir(fpJoin), 0755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(fpJoin, data, 0644); err != nil {
		panic(err)
	}
}

// DeleteAll is used to delete the file or folder specified and all its contents. The path should be relative to
// the data folder.
func (t *Transaction) DeleteAll(path string) {
	// Delete the folder from the commit data folder if it exists.
	if err := os.RemoveAll(filepath.Join(t.getTransactionFolder(), "d", path)); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	}

	// Write the journal action.
	t.writeJournalAction(journalAction{
		T:  actionTypeDeleteAll,
		P1: path,
	})
}

// Delete is used to delete the file specified. The path should be relative to the data folder.
func (t *Transaction) Delete(path string) {
	// Delete the file from the commit data folder if it exists.
	if err := os.Remove(filepath.Join(t.getTransactionFolder(), "d", path)); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	}

	// Write the journal action.
	t.writeJournalAction(journalAction{
		T:  actionTypeDelete,
		P1: path,
	})
}

// Rename is used to rename a file. The path should be relative to the data folder.
func (t *Transaction) Rename(path, newPath string) {
	// Rename the file in the commit data folder if it exists.
	txFolder := t.getTransactionFolder()
	p1 := filepath.Join(txFolder, "d", path)
	p2 := filepath.Join(txFolder, "d", newPath)
	if err := os.Rename(p1, p2); err != nil {
		// If the error is not that the file doesn't exist, panic.
		if !os.IsNotExist(err) {
			panic(err)
		}

		// Attempt a copy to the transaction folder.
		s := filepath.SplitList(p2)
		if s[len(s)-1] == "" {
			s = s[:len(s)-1]
		}
		if err := os.MkdirAll(filepath.Join(s[:len(s)-1]...), 0755); err != nil {
			panic(err)
		}
		_ = cp.Copy(filepath.Join(t.dataPath, path), p2)
	}

	// Write the journal action.
	t.writeJournalAction(journalAction{
		T:  actionTypeRename,
		P1: path,
		P2: newPath,
	})
}

// Mkdir is used to create a folder. The path should be relative to the data folder.
func (t *Transaction) Mkdir(path string) {
	t.writeJournalAction(journalAction{
		T:  actionTypeMkdir,
		P1: path,
	})
}

// MkdirAll is used to create a folder and all its parents. The path should be relative to the data folder.
func (t *Transaction) MkdirAll(path string) {
	t.writeJournalAction(journalAction{
		T:  actionTypeMkdirAll,
		P1: path,
	})
}

// ReadFile is used to read a file from the data folder. The path should be relative to the data folder.
func (t *Transaction) ReadFile(path string) ([]byte, error) {
	fp := filepath.Join(t.getTransactionFolder(), "d", path)
	b, err := os.ReadFile(fp)
	if err == nil {
		return b, nil
	} else {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return os.ReadFile(filepath.Join(t.dataPath, path))
}

// Handles the transaction commit. The data to prepare a commit has already been written.
func (t *Transaction) handleCommit(txFp string) error {
	// Go through all the journal actions.
	for action := t.journalActionsStart; action != nil; action = action.next {
		// Do the action.
		switch action.T {
		case actionTypeCreate:
			// Get the transaction file we are renaming.
			txFile := filepath.Join(txFp, "d", action.P1)
			s, err := os.Stat(txFile)
			if err != nil || s.IsDir() {
				// If something changed, ignore this. Probably stale.
				continue
			}

			// Get the destination file.
			destFile := filepath.Join(t.dataPath, action.P1)

			// Do the rename.
			if err := os.Rename(txFile, destFile); err != nil {
				if os.IsExist(err) || os.IsNotExist(err) {
					// If something changed, ignore this. Probably stale.
					continue
				}

				return err
			}
		case actionTypeDeleteAll:
			// Get the file or folder in question.
			txFile := filepath.Join(t.dataPath, action.P1)

			// Delete it.
			if err := os.RemoveAll(txFile); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			}
		case actionTypeDelete:
			// Get the file or folder in question.
			txFile := filepath.Join(t.dataPath, action.P1)

			// Delete it.
			if err := os.Remove(txFile); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			}
		case actionTypeRename:
			// Do the rename.
			txFile1 := filepath.Join(t.dataPath, action.P1)
			txFile2 := filepath.Join(t.dataPath, action.P2)
			if err := os.Rename(txFile1, txFile2); err != nil {
				if os.IsExist(err) || os.IsNotExist(err) {
					// If something changed, ignore this. Probably stale.
					continue
				}

				return err
			}
		case actionTypeMkdir:
			// Do the mkdir.
			txFile := filepath.Join(t.dataPath, action.P1)
			if err := os.Mkdir(txFile, 0755); err != nil {
				if os.IsExist(err) || os.IsNotExist(err) {
					// If something changed, ignore this. Probably stale.
					continue
				}

				return err
			}
		case actionTypeMkdirAll:
			// Do the recursive mkdir.
			txFile := filepath.Join(t.dataPath, action.P1)
			if err := os.MkdirAll(txFile, 0755); err != nil {
				if os.IsExist(err) || os.IsNotExist(err) {
					// If something changed, ignore this. Probably stale.
					continue
				}

				return err
			}
		}

		// Delete the journal action.
		fp := filepath.Join(txFp, strconv.Itoa(action.num)+".J")
		if err := os.Remove(fp); err != nil {
			return err
		}
	}

	// No errors!
	return nil
}
