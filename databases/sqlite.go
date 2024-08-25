package databases

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/sj14/dbbench/benchmark"
)

// SQLite implements the bencher interface.
type SQLite struct {
	db *sql.DB
}

var (
	dbPath    string
	dbCreated bool // DB file was created by dbbench
)

// NewSQLite retruns a new SQLite bencher.
func NewSQLite(path string) *SQLite {
	dbPath = path

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// We will create the database file.
		dbCreated = true
	}

	// Automatically creates the DB file if it doesn't exist yet.
	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?cache=shared", path))
	if err != nil {
		log.Fatalf("failed to open connection: %v\n", err)
	}

	db.SetMaxOpenConns(1)
	p := &SQLite{db: db}
	return p
}

// Benchmarks returns the individual benchmark statements for sqlite.
func (m *SQLite) Benchmarks() []benchmark.Benchmark {
	return []benchmark.Benchmark{
		{Name: "inserts", Type: benchmark.TypeLoop, Stmt: "INSERT INTO dbbench_simple (id, balance) VALUES( {{.Iter}}, {{call .RandInt64}});"},
		{Name: "selects", Type: benchmark.TypeLoop, Stmt: "SELECT * FROM dbbench_simple WHERE id = {{.Iter}};"},
		{Name: "updates", Type: benchmark.TypeLoop, Stmt: "UPDATE dbbench_simple SET balance = {{call .RandInt64}} WHERE id = {{.Iter}};"},
		{Name: "deletes", Type: benchmark.TypeLoop, Stmt: "DELETE FROM dbbench_simple WHERE id = {{.Iter}};"},
		// {"relation_insert0", benchmark.TypeLoop, "INSERT INTO dbbench_relational_one (oid, balance_one) VALUES( {{.Iter}}, {{call .RandInt64}});"},
		// {"relation_insert1", benchmark.TypeLoop, "INSERT INTO dbbench_relational_two (relation, balance_two) VALUES( {{.Iter}}, {{call .RandInt64}});"},
		// {"relation_select", benchmark.TypeLoop, "SELECT * FROM dbbench_relational_two INNER JOIN dbbench_relational_one ON dbbench_relational_one.oid = relation WHERE relation = {{.Iter}};"},
		// {"relation_delete1", benchmark.TypeLoop, "DELETE FROM dbbench_relational_two WHERE relation = {{.Iter}};"},
		// {"relation_delete0", benchmark.TypeLoop, "DELETE FROM dbbench_relational_one WHERE oid = {{.Iter}};"},
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
	if _, err := m.db.Exec("CREATE TABLE IF NOT EXISTS dbbench_relational_two (balance_two DECIMAL, relation INT PRIMARY KEY, FOREIGN KEY(relation) REFERENCES dbbench_relational_one(oid));"); err != nil {
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

	// The DB file existed before, don't remove it.
	if !dbCreated {
		return
	}

	if err := os.Remove(dbPath); err != nil {
		log.Printf("not able to delete created database file: %v\n", err)
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
