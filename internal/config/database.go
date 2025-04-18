package config

import (
	"context"
	"database/sql"
	"fmt"
	"os"
)

func setupTables(db *sql.DB) {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS cronjobs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			external_id varchar(255) NOT NULL,
			name varchar(255) NOT NULL,
			expression varchar(50) NOT NULL,
			created_at TIMESTAMP DEFAULT current_timestamp,
			next_execution TIMESTAMP NOT NULL,
			enabled BOOLEAN DEFAULT TRUE,
			url_to_trigger text NOT NULL,
			is_executing BOOLEAN DEFAULT FALSE
		);

		CREATE TABLE IF NOT EXISTS cronjobs_logs (
			id integer PRIMARY KEY AUTOINCREMENT,
			external_id varchar(255) NOT NULL,
			data text,
			created_at TIMESTAMP DEFAULT current_timestamp
		);
	`)
	if err != nil {
		panic(fmt.Sprintf("failed to create table: %v", err))
	}

	fmt.Println("Table created.")
}

func InitDB() *sql.DB {
	url := os.Getenv("TURSO_DATABASE_URL")
	token := os.Getenv("TURSO_AUTH_TOKEN")
	dsn := fmt.Sprintf("%s?authToken=%s", url, token)

	db, err := sql.Open("libsql", dsn)
	if err != nil {
		panic(fmt.Sprintf("failed to connect: %v", err))
	}

	setupTables(db)

	return db
}
