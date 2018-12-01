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
		{"inserts", Loop, "INSERT INTO dbbench_simple (id, balance) VALUES( {{.Iter}}, {{call .RandInt63}});"},
		{"selects", Loop, "SELECT * FROM dbbench_simple WHERE id = {{.Iter}};"},
		{"updates", Loop, "UPDATE dbbench_simple SET balance = {{call .RandInt63}} WHERE id = {{.Iter}};"},
		{"deletes", Loop, "DELETE FROM dbbench_simple WHERE id = {{.Iter}};"},
		{"relation_insert0", Loop, "INSERT INTO dbbench_relational_one (oid, balance_one) VALUES( {{.Iter}}, {{call .RandInt63}});"},
		{"relation_insert1", Loop, "INSERT INTO dbbench_relational_two (relation, balance_two) VALUES( {{.Iter}}, {{call .RandInt63}});"},
		{"relation_select", Loop, "SELECT * FROM dbbench_relational_two INNER JOIN dbbench_relational_one ON dbbench_relational_one.oid = relation WHERE relation = {{.Iter}};"},
		{"relation_delete1", Loop, "DELETE FROM dbbench_relational_two WHERE relation = {{.Iter}};"},
		{"relation_delete0", Loop, "DELETE FROM dbbench_relational_one WHERE oid = {{.Iter}};"},
	}
}

// Setup initializes the database for the benchmark.
func (m *SQLite) Setup() {
	if _, err := m.db.Exec("CREATE TABLE IF NOT EXISTS dbbench_simple (id INT PRIMARY KEY, balance DECIMAL);"); err != nil {
		log.Fatalf("failed to create table dbbench_simple: %v\n", err)
	}
	if _, err := m.db.Exec("CREATE TABLE IF NOT EXISTS dbbench_relational_one (oid INT PRIMARY KEY, balance_one DECIMAL);"); err != nil {
		log.Fatalf("failed to create table dbbench_relational_one: %v\n", err)
	}
	if _, err := m.db.Exec("CREATE TABLE IF NOT EXISTS dbbench_relational_two (balance_two DECIMAL, relation INT, FOREIGN KEY(relation) REFERENCES dbbench_relational_one(oid));"); err != nil {
		log.Fatalf("failed to create table dbbench_relational_two: %v\n", err)
	}
	if _, err := m.db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Fatalf("failed to enabled foreign keys: %v\n", err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (m *SQLite) Cleanup() {
	if _, err := m.db.Exec("DROP TABLE dbbench_simple"); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	if _, err := m.db.Exec("DROP TABLE dbbench_relational_two"); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	if _, err := m.db.Exec("DROP TABLE dbbench_relational_one"); err != nil {
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
	//  driver has no support for results
	_, err := m.db.Exec(stmt)
	if err != nil {
		log.Printf("%v failed: %v", stmt, err)
	}
}
