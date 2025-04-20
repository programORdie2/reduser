package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	// run migrations
	stmts := []string{
		`PRAGMA foreign_keys = ON;`,
		`CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT NOT NULL UNIQUE,
            password_hash TEXT NOT NULL
        );`,
		`CREATE TABLE IF NOT EXISTS projects (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER NOT NULL,
            name TEXT NOT NULL,
            token TEXT NOT NULL UNIQUE,
            FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
        );`,
		`CREATE TABLE IF NOT EXISTS tables (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            project_id INTEGER NOT NULL,
            user_id INTEGER NOT NULL,
            name TEXT NOT NULL,
            FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
            FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
            UNIQUE(project_id, user_id, name)
        );`,
		`CREATE TABLE IF NOT EXISTS variables (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            table_id INTEGER NOT NULL,
            user_id INTEGER NOT NULL,
            name TEXT NOT NULL,
            value TEXT,
            type TEXT NOT NULL CHECK (type IN ('string','int','float','bool')),
            FOREIGN KEY(table_id) REFERENCES tables(id) ON DELETE CASCADE,
            FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
            UNIQUE(table_id, user_id, name)
        );`,
	}

	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			log.Fatalf("migration error: %v\nstmt: %s", err, s)
		}
	}

	return db
}
