package storage

import (
	"context"
	"database/sql"
)

func createTables(ctx context.Context, db *sql.DB) error {
	const (
		usersTable = `
	CREATE TABLE IF NOT EXISTS users(
		login TEXT PRIMARY KEY, 
		password TEXT
	);`

		expressionsTable = `
	CREATE TABLE IF NOT EXISTS expressions(
		expression_id INTEGER PRIMARY KEY AUTOINCREMENT,
		status TEXT,
		result REAL,
		binary_tree_bytes BLOB NOT NULL,
		login TEXT,

		FOREIGN KEY (login) REFERENCES users (login)
	);`
		tasksTable = `
	CREATE TABLE IF NOT EXISTS tasks(
		task_id INTEGER PRIMARY KEY AUTOINCREMENT,
		status TEXT,
		arg1 REAL,
		arg2 REAL,
		operation TEXT,
		operation_time INTEGER, --наносекунды
		expression_id INTEGER,

		FOREIGN KEY (expression_id) REFERENCES expressions (expression_id)
	);`
	)

	if _, err := db.ExecContext(ctx, usersTable); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, expressionsTable); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, tasksTable); err != nil {
		return err
	}

	return nil
}