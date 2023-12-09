// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package acid

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
)

// Transaction is used to define a ACID transaction. These transactions are used to ensure that the database is always in a consistent state.
// For example, if you do a IO operation and the system crashes, without this the underlying data could be corrupted. With this, the system will
// self-heal and ensure that the data is always in a consistent state. You can create many transactions, but the transaction object itself
// is not thread safe. You must call Commit or Rollback on the transaction when you are done with it to prevent resource leaks. When you
// do actions within the transactions, you should strongly consider the order of the actions. For example, if you are deleting a file, you do
// not want to delete the path under it first.
type Transaction struct {
	// TransactionID is the ID of the transaction.
	TransactionID string

	dataPath            string
	journalActionsStart *journalAction
	journalActionsEnd   *journalAction
	count               int
	folderMade          bool
	done                bool
}

// ErrAlreadyCommitted is used to define the error when the transaction has already been committed.
var ErrAlreadyCommitted = errors.New("transaction already committed")

func (t *Transaction) handleDone() error {
	if t.done {
		return ErrAlreadyCommitted
	}
	return nil
}

func (t *Transaction) getTransactionFolder() string {
	fp := filepath.Join(t.dataPath, "transactions", t.TransactionID)

	if t.folderMade {
		// Return early since we do not need to do any FS operations.
		return fp
	}

	if err := os.Mkdir(fp, 0755); err != nil {
		// Try MkdirAll in case the parent directory does not exist.
		if err := os.MkdirAll(fp, 0755); err != nil {
			// In this case panic because there has been irrecoverable damage to the database.
			panic(err)
		}
	}

	t.folderMade = true
	return fp
}

// Rollback is used to rollback the transaction. This will undo all the changes made in the transaction.
// Returns ErrAlreadyCommitted if the transaction has already been committed.
func (t *Transaction) Rollback() error {
	if err := t.handleDone(); err != nil {
		return err
	}

	if t.folderMade {
		// Remove the transaction folder.
		if err := os.RemoveAll(t.getTransactionFolder()); err != nil {
			return err
		}
	}

	t.done = true
	return nil
}

// Commit is used to commit the transaction. This will commit all the changes made in the transaction.
// Returns ErrAlreadyCommitted if the transaction has already been committed.
func (t *Transaction) Commit() error {
	// Handle the done state.
	if err := t.handleDone(); err != nil {
		return err
	}

	// Write the special 'C' file to indicate that the transaction has been committed.
	txFp := t.getTransactionFolder()
	commitFp := filepath.Join(txFp, "C")
	if err := os.WriteFile(commitFp, []byte{}, 0644); err != nil {
		return err
	}

	// Do the commit.
	if err := t.handleCommit(txFp); err != nil {
		return err
	}

	// Delete the C file.
	if err := os.Remove(commitFp); err != nil {
		return err
	}

	// Delete the transaction folder.
	if err := os.RemoveAll(txFp); err != nil {
		return err
	}

	// Mark the transaction as done.
	t.done = true
	return nil
}

// New is used to create a new transaction.
func New(path string) *Transaction {
	return &Transaction{
		TransactionID: uuid.NewString(),
		dataPath:      path,
	}
}

var jRegex = regexp.MustCompile(`^(\d+)\.J$`)

// RecoverFailedTransaction is used to figure out what to do with a failed transaction. If the transaction is nil,
// the transaction was never committed. If the transaction is not nil, the transaction was committed and the commit
// should be tried again.
func RecoverFailedTransaction(path, transactionId string) *Transaction {
	// Build a object with the broken transaction.
	tx := &Transaction{
		TransactionID: transactionId,
		dataPath:      path,
		folderMade:    true,
	}

	// Check if the transaction was committed.
	txFp := tx.getTransactionFolder()
	commitFp := filepath.Join(txFp, "C")
	_, err := os.Stat(commitFp)
	if err != nil {
		if os.IsNotExist(err) {
			// The transaction was not committed. Delete the transaction folder.
			if err := os.RemoveAll(txFp); err != nil {
				panic(err)
			}
			return nil
		}

		// Another error occurred.
		panic(err)
	}

	// Get all the files in the directory.
	dir, err := os.ReadDir(txFp)
	if err != nil {
		panic(err)
	}

	// Re-create the journal actions.
	journalNumber := 0
	for _, d := range dir {
		// Check if the file is a journal action.
		name := d.Name()
		matches := jRegex.FindStringSubmatch(name)
		if matches == nil {
			continue
		}

		// Get the journal number.
		journalNumberStr := matches[1]
		journalNumberThis, err := strconv.Atoi(journalNumberStr)
		if err != nil {
			panic(err)
		}
		if journalNumberThis > journalNumber {
			journalNumber = journalNumberThis
		}

		// Read from the file.
		b, err := os.ReadFile(filepath.Join(txFp, name))
		if err != nil {
			panic(err)
		}
		var j journalAction
		if err := msgpack.Unmarshal(b, &j); err != nil {
			panic(err)
		}

		// Set the number.
		j.num = journalNumberThis

		// Add it to the transaction.
		if tx.journalActionsStart == nil {
			tx.journalActionsStart = &j
			tx.journalActionsEnd = &j
		} else {
			tx.journalActionsEnd.next = &j
			tx.journalActionsEnd = &j
		}
	}
	tx.count = journalNumber + 1

	// Return the transaction.
	return tx
}
