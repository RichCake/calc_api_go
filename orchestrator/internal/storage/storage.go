package storage

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func createTables(ctx context.Context, db *sql.DB) error {
	const (
		usersTable = `
	CREATE TABLE IF NOT EXISTS users(
		user_id INTEGER PRIMARY KEY AUTOINCREMENT,
		login TEXT UNIQUE, 
		password TEXT
	);`

		expressionsTable = `
	CREATE TABLE IF NOT EXISTS expressions(
		expression_id INTEGER PRIMARY KEY AUTOINCREMENT,
		status TEXT,
		result REAL,
		binary_tree_bytes TEXT NOT NULL,
		user_id INTEGER,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,

		FOREIGN KEY (user_id) REFERENCES users (user_id)
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

type Storage struct {
	db *sql.DB
}

func NewStorage(for_tests bool) *Storage {
	ctx := context.TODO()
	
	dbDir := filepath.Join("orchestrator", "storage")
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		os.Mkdir(dbDir, 0777)
	}
	dbPath := filepath.Join(dbDir, "store.db")

	if for_tests {
		dbPath = ":memory:"
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		panic(err)
	}

	if err = createTables(ctx, db); err != nil {
		panic(err)
	}
	return &Storage{
		db: db,
	}
}

func (s *Storage) Close() error {
	return s.db.Close()
}