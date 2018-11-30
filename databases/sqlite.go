package databases

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

// SQLite implements the bencher implementation.
type SQLite struct {
	db *sql.DB
}

var dbPath string

// NewSQLite retruns a new mysql bencher.
func NewSQLite(path string) *SQLite {
	dbPath = path

	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?cache=shared", path))
	if err != nil {
		log.Fatalf("failed to open connection: %v\n", err)
	}

	db.SetMaxOpenConns(1)
	p := &SQLite{db: db}
	return p
}

// Benchmarks returns the individual benchmark statements for sqlite.
func (m *SQLite) Benchmarks() []Benchmark {
	return []Benchmark{
		{"inserts", Loop, "INSERT INTO accounts (id, balance) VALUES( {{.Iter}}, {{call .RandInt63}});"},
		{"selects", Loop, "SELECT * FROM accounts WHERE id = {{.Iter}};"},
		{"updates", Loop, "UPDATE accounts SET balance = {{call .RandInt63}} WHERE id = {{.Iter}};"},
		{"deletes", Loop, "DELETE FROM accounts WHERE id = {{.Iter}};"},
	}
}

// Setup initializes the database for the benchmark.
func (m *SQLite) Setup() {
	if _, err := m.db.Exec("CREATE TABLE IF NOT EXISTS accounts (id INT PRIMARY KEY, balance DECIMAL);"); err != nil {
		log.Fatalf("failed to create table: %v\n", err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (m *SQLite) Cleanup() {
	if _, err := m.db.Exec("DROP TABLE accounts"); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	if err := m.db.Close(); err != nil {
		log.Printf("failed to close connection: %v", err)
	}

	// TODO: Attention: should we really remove the file?
	if err := os.Remove(dbPath); err != nil {
		log.Printf("not able to delete old database file: %v\n", err)
	}
}

// Exec executes the given statement on the database.
func (m *SQLite) Exec(stmt string) {
	_, err := m.db.Exec(stmt)
	if err != nil {
		log.Printf("%v failed: %v", stmt, err)
	}
}
