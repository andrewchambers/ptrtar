package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/mattn/go-sqlite3"
)

type CreateCache struct {
	db *sql.DB
	tx *sql.Tx

	hasPtrStmt *sql.Stmt
	addPtrStmt *sql.Stmt
}

func IsRetryableSqliteError(err error) bool {
	if err == nil {
		return false
	}

	switch err := err.(type) {
	case *sqlite3.Error:
		if err.Code == sqlite3.ErrBusy {
			return true
		}
		if err.Code == sqlite3.ErrLocked {
			return true
		}
	}
	return false
}

func RetryableTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	return RetryableTxWithOptions(ctx, db, &sql.TxOptions{Isolation: sql.LevelSerializable}, fn)
}

// Run fn inside a transaction, on error abort the transaction, unless the
// error is a retryable sqlite error.
//
func RetryableTxWithOptions(ctx context.Context, db *sql.DB, opts *sql.TxOptions, fn func(*sql.Tx) error) error {

	do := func() (err error) {
		tx, txErr := db.BeginTx(ctx, opts)
		if txErr != nil {
			return txErr
		}

		defer func() {
			r := recover()
			if r != nil && err == nil {
				err = errors.New("panic")
			}

			if err == nil {
				err = tx.Commit()
			} else {
				_ = tx.Rollback()
			}

			if r != nil {
				panic(r)
			}

		}()

		return fn(tx)
	}

	retried := false
	for {
		err := do()
		if err == nil {
			return nil
		}

		if !IsRetryableSqliteError(err) {
			return err
		}

		if retried {
			time.Sleep(100 * time.Millisecond)
		}
		retried = true
	}
}

const PtrTarCacheVersion = 0

func OpenCache(dbPath string) (*CreateCache, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	err = RetryableTx(context.Background(), db, func(tx *sql.Tx) error {
		var version uint32

		err = tx.QueryRow("SELECT Value from Meta WHERE Name='PtrTarCacheVersion';").Scan(&version)
		if err == nil {
			if version != PtrTarCacheVersion {
				return fmt.Errorf("write cache %s version %d differs from your current software version %d. (It is safe to delete the old cache file manually).", dbPath, version, PtrTarCacheVersion)
			}
		}

		_, err = tx.Exec("CREATE TABLE IF NOT EXISTS Files (FullPath, ModTimeNano, ChangeTimeNano, Size, CachedPtr, UNIQUE(FullPath));")
		if err != nil {
			return err
		}

		_, err = tx.Exec("CREATE TABLE IF NOT EXISTS Meta (Name UNIQUE, Value);")
		if err != nil {
			return err
		}

		_, err = tx.Exec("INSERT OR IGNORE INTO Meta (Name, Value) Values ('PtrTarCacheVersion', ?);", PtrTarCacheVersion)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	hasPtrStmt, err := tx.Prepare("SELECT CachedPtr FROM Files WHERE FullPath = ? AND ModTimeNano = ? AND ChangeTimeNano = ? AND SIZE = ?;")
	if err != nil {
		return nil, err
	}
	addPtrStmt, err := tx.Prepare("INSERT OR REPLACE INTO Files (FullPath, ModTimeNano, ChangeTimeNano, Size, CachedPtr) VALUES (?,?,?,?,?);")
	if err != nil {
		return nil, err
	}

	return &CreateCache{
		db:         db,
		tx:         tx,
		hasPtrStmt: hasPtrStmt,
		addPtrStmt: addPtrStmt,
	}, nil
}

func (wcache *CreateCache) HasPtr(fullPath string, modTime, changeTime time.Time, size int64) ([]byte, bool, error) {
	var ptr []byte

	err := wcache.hasPtrStmt.QueryRow(fullPath, modTime.UnixNano(), changeTime.UnixNano(), size).Scan(&ptr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, err
	}
	return ptr, true, nil
}

func (wcache *CreateCache) AddPtr(fullPath string, modTime, changeTime time.Time, size int64, ptr []byte) error {
	_, err := wcache.addPtrStmt.Exec(fullPath, modTime.UnixNano(), changeTime.UnixNano(), size, ptr)
	if err != nil {
		return err
	}
	return nil
}

func (wcache *CreateCache) Close() error {
	if wcache.tx != nil {
		_ = wcache.hasPtrStmt.Close()
		_ = wcache.addPtrStmt.Close()
		_ = wcache.tx.Commit()
		_ = wcache.db.Close()
	}
	return nil
}
