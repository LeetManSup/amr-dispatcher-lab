package store

import (
	"context"
	"database/sql"
	"time"
)

func execMany(db *sql.DB, query string, callback func(*sql.Stmt) error) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()
	if err := callback(stmt); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func nullableTime(value *time.Time) interface{} {
	if value == nil {
		return nil
	}
	return value.Format(time.RFC3339Nano)
}
