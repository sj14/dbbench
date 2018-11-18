package cockroach

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // pq is the postgres/cockroach db driver
)

// Cockroach implements the bencher implementation.
type Cockroach struct {
	db *sql.DB
}

// New returns a new cockroach bencher.
func New() *Cockroach {
	db, err := sql.Open("postgres", "host=127.0.0.1 port=26257 user=root password='' sslmode=disable")
	if err != nil {
		log.Fatalf("failed to open connection: %v\n", err)
	}

	c := &Cockroach{db: db}
	return c
}

// Benchmarks returns the individual benchmark functions for the cockroach db.
func (p *Cockroach) Benchmarks() []func(int) string {
	return []func(int) string{p.inserts, p.updates, p.selects, p.deletes}

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
		log.Fatalln(err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (p *Cockroach) Cleanup() {
	if _, err := p.db.Exec("DROP TABLE dbbench.accounts"); err != nil {
		log.Fatalf("failed to drop table: %v\n", err)
	}
	if _, err := p.db.Exec("DROP DATABASE dbbench"); err != nil {
		log.Fatalf("failed to drop database: %v\n", err)
	}
	p.db.Close()
}

func (p *Cockroach) inserts(iterations int) string {
	const q = "INSERT INTO dbbench.accounts VALUES($1, $2);"
	for i := 0; i < iterations; i++ {
		if _, err := p.db.Exec(q, i, i); err != nil {
			log.Fatalf("failed to insert: %v\n", err)
		}
	}
	return "inserts"
}

func (p *Cockroach) selects(iterations int) string {
	const q = "SELECT * FROM dbbench.accounts WHERE id = $1;"
	for i := 0; i < iterations; i++ {
		if _, err := p.db.Exec(q, i); err != nil {
			log.Fatalf("failed to select: %v\n", err)
		}
	}
	return "selects"
}

func (p *Cockroach) updates(iterations int) string {
	const q = "UPDATE dbbench.accounts SET balance = $1 WHERE id = $2;"
	for i := 0; i < iterations; i++ {
		if _, err := p.db.Exec(q, i, i); err != nil {
			log.Fatalf("failed to update: %v\n", err)
		}
	}
	return "updates"
}

func (p *Cockroach) deletes(iterations int) string {
	const q = "DELETE FROM dbbench.accounts WHERE id = $1"
	for i := 0; i < iterations; i++ {
		if _, err := p.db.Exec(q, i); err != nil {
			log.Fatalf("failed to delete: %v\n", err)
		}
	}
	return "deletes"
}
