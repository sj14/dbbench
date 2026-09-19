package databases

import (
	"database/sql"
	"log"
	"os"

	"github.com/sj14/dbbench/benchmark"
)

// Turso implements the bencher interface using the Tursogo driver.
type Turso struct {
	db      *sql.DB
	path    string
	created bool
}

// NewTurso returns a new Turso bencher.
func NewTurso(path string, maxconns int) *Turso {
	_, err := os.Stat(path)
	created := os.IsNotExist(err)

	db, err := sql.Open("turso", path)
	if err != nil {
		log.Fatalf("failed to open connection: %v\n", err)
	}
	db.SetMaxOpenConns(maxconns)

	return &Turso{db: db, path: path, created: created}
}

// Benchmarks returns the individual benchmark statements for Turso.
func (t *Turso) Benchmarks() []benchmark.Benchmark {
	return []benchmark.Benchmark{
		{Name: "inserts", Type: benchmark.TypeLoop, Stmt: "INSERT INTO dbbench_simple (id, balance) VALUES( {{.Iter}}, {{call .RandInt64}});"},
		{Name: "selects", Type: benchmark.TypeLoop, Stmt: "SELECT * FROM dbbench_simple WHERE id = {{.Iter}};"},
		{Name: "updates", Type: benchmark.TypeLoop, Stmt: "UPDATE dbbench_simple SET balance = {{call .RandInt64}} WHERE id = {{.Iter}};"},
		{Name: "deletes", Type: benchmark.TypeLoop, Stmt: "DELETE FROM dbbench_simple WHERE id = {{.Iter}};"},
	}
}

// Setup initializes the database for the benchmark.
func (t *Turso) Setup() {
	var existing int
	if err := t.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name IN ('dbbench_simple', 'dbbench_relational_one', 'dbbench_relational_two')").Scan(&existing); err != nil {
		log.Fatalf("failed to inspect existing tables: %v\n", err)
	}
	if existing != 0 {
		log.Fatalf("benchmark tables already exist; use --clean to remove stale benchmark data")
	}

	if _, err := t.db.Exec("CREATE TABLE dbbench_simple (id INT PRIMARY KEY, balance DECIMAL);"); err != nil {
		log.Fatalf("failed to create table dbbench_simple (use --clean to remove stale benchmark data): %v\n", err)
	}
	if _, err := t.db.Exec("CREATE TABLE dbbench_relational_one (oid INT PRIMARY KEY, balance_one DECIMAL);"); err != nil {
		log.Fatalf("failed to create table dbbench_relational_one: %v\n", err)
	}
	if _, err := t.db.Exec("CREATE TABLE dbbench_relational_two (balance_two DECIMAL, relation INT PRIMARY KEY, FOREIGN KEY(relation) REFERENCES dbbench_relational_one(oid));"); err != nil {
		log.Fatalf("failed to create table dbbench_relational_two: %v\n", err)
	}
	if _, err := t.db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Fatalf("failed to enable foreign keys: %v\n", err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (t *Turso) Cleanup() {
	if _, err := t.db.Exec("DROP TABLE dbbench_simple"); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	if _, err := t.db.Exec("DROP TABLE dbbench_relational_two"); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	if _, err := t.db.Exec("DROP TABLE dbbench_relational_one"); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	if err := t.db.Close(); err != nil {
		log.Printf("failed to close connection: %v", err)
	}

	if !t.created {
		return
	}
	for _, path := range []string{t.path, t.path + "-wal"} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			log.Printf("not able to delete created database file %s: %v\n", path, err)
		}
	}
}

// Exec executes the given statement on the database.
func (t *Turso) Exec(stmt string) error {
	_, err := t.db.Exec(stmt)
	return err
}
