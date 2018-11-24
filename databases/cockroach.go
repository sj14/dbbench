package databases

import (
	"database/sql"
	"fmt"
	"log"
)

// Cockroach implements the bencher implementation.
type Cockroach struct {
	db *sql.DB
}

// NewCockroach returns a new cockroach bencher.
func NewCockroach(host string, port int, user, password string, maxOpenConns int) *Cockroach {
	if port == 0 {
		port = 26257
	}

	dataSourceName := fmt.Sprintf("host=%v port=%v user='%v' password='%v' sslmode=disable", host, port, user, password)

	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatalf("failed to open connection: %v\n", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	return &Cockroach{db: db}
}

// Benchmarks returns the individual benchmark functions for the cockroach db.
func (p *Cockroach) Benchmarks() []Benchmark {
	return []Benchmark{
		{"inserts", Loop, p.inserts},
		{"updates", Loop, p.updates},
		{"selects", Loop, p.selects},
		{"deletes", Loop, p.deletes},
	}
}

// Setup initializes the database for the benchmark.
func (p *Cockroach) Setup() {
	if _, err := p.db.Exec("CREATE DATABASE IF NOT EXISTS dbbench"); err != nil {
		log.Fatalf("failed to create database: %v\n", err)
	}
	if _, err := p.db.Exec("CREATE TABLE IF NOT EXISTS dbbench.accounts (id INT PRIMARY KEY, balance DECIMAL);"); err != nil {
		log.Fatalf("failed to create table: %v\n", err)
	}
	if _, err := p.db.Exec("TRUNCATE dbbench.accounts;"); err != nil {
		log.Fatalf("failed to truncate table: %v\n", err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (p *Cockroach) Cleanup() {
	if _, err := p.db.Exec("DROP TABLE dbbench.accounts"); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	if _, err := p.db.Exec("DROP DATABASE dbbench"); err != nil {
		log.Printf("failed to drop database: %v\n", err)
	}
	if err := p.db.Close(); err != nil {
		log.Printf("failed to close connection: %v", err)
	}
}

// Exec executes the given statement on the database.
func (p *Cockroach) Exec(stmt string) {
	result, err := p.db.Exec(stmt)
	mustExec(result, err, stmt)
}

func (p *Cockroach) inserts(i int) {
	const q = "INSERT INTO dbbench.accounts VALUES($1, $2);"
	if _, err := p.db.Exec(q, i, i); err != nil {
		log.Fatalf("failed to insert: %v\n", err)
	}
}

func (p *Cockroach) selects(i int) {
	const q = "SELECT * FROM dbbench.accounts WHERE id = $1;"
	if _, err := p.db.Exec(q, i); err != nil {
		log.Fatalf("failed to select: %v\n", err)
	}
}

func (p *Cockroach) updates(i int) {
	const q = "UPDATE dbbench.accounts SET balance = $1 WHERE id = $2;"
	if _, err := p.db.Exec(q, i, i); err != nil {
		log.Fatalf("failed to update: %v\n", err)
	}
}

func (p *Cockroach) deletes(i int) {
	const q = "DELETE FROM dbbench.accounts WHERE id = $1"
	if _, err := p.db.Exec(q, i); err != nil {
		log.Fatalf("failed to delete: %v\n", err)
	}
}
