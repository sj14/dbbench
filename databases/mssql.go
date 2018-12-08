package databases

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"

	"github.com/sj14/dbbench/benchmark"
)

// MSSQL implements the bencher implementation.
type MSSQL struct {
	db *sql.DB
}

// NewMSSQL returns a new MS SQL bencher.
func NewMSSQL(host string, port int, user, password string, maxOpenConns int) *MSSQL {
	if port == 0 {
		port = 1433
	}

	u := &url.URL{
		Scheme: "sqlserver",
		User:   url.UserPassword(user, password),
		Host:   fmt.Sprintf("%s:%d", host, port),
		// Path:  instance, // TODO: when connecting to an instance instead of a port
	}

	db, err := sql.Open("sqlserver", u.String())
	if err != nil {
		log.Fatalf("failed to open connection: %v\n", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	p := &MSSQL{db: db}
	return p
}

// Benchmarks returns the individual benchmark functions for the mysql db.
func (m *MSSQL) Benchmarks() []benchmark.Benchmark {
	log.Fatal("no built-in benchmarks for MS SQL available yet, use your own script")
	return []benchmark.Benchmark{}
}

// Setup initializes the database for the benchmark.
func (m *MSSQL) Setup() {
}

// Cleanup removes all remaining benchmarking data.
func (m *MSSQL) Cleanup() {
}

// Exec executes the given statement on the database.
func (m *MSSQL) Exec(stmt string) {
	_, err := m.db.Exec(stmt)
	if err != nil {
		log.Printf("%v failed: %v", stmt, err)
	}
}
