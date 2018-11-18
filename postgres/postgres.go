package postgres

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // pq is the postgres db driver
)

// Postgres implements the bencher implementation.
type Postgres struct {
	db *sql.DB
}

// New returns a new postgres bencher.
func New(host string, port int, user, password string) *Postgres {
	dataSourceName := fmt.Sprintf("host=%v port=%v user='%v' password='%v' sslmode=disable", host, port, user, password)

	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatalf("failed to open connection: %v\n", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	p := &Postgres{db: db}
	return p
}

// Benchmarks returns the individual benchmark functions for the postgres db.
func (p *Postgres) Benchmarks() []func(int, int) string {
	return []func(int, int) string{p.inserts, p.updates, p.selects, p.deletes}
}

// Setup initializes the database for the benchmark.
func (p *Postgres) Setup() {
	if _, err := p.db.Exec("CREATE SCHEMA IF NOT EXISTS dbbench"); err != nil {
		log.Fatalf("failed to create schema: %v\n", err)
	}
	if _, err := p.db.Exec("CREATE TABLE IF NOT EXISTS dbbench.accounts (id INT PRIMARY KEY, balance DECIMAL);"); err != nil {
		log.Fatalf("failed to cfreate table: %v\n", err)
	}
	if _, err := p.db.Exec("TRUNCATE dbbench.accounts;"); err != nil {
		log.Fatalf("failed to truncate table: %v\n", err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (p *Postgres) Cleanup() {
	if _, err := p.db.Exec("DROP TABLE dbbench.accounts"); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	if _, err := p.db.Exec("DROP SCHEMA dbbench"); err != nil {
		log.Printf("failed drop schema: %v\n", err)
	}
	if err := p.db.Close(); err != nil {
		log.Printf("failed to close connection: %v", err)
	}
}

func (p *Postgres) inserts(from, to int) string {
	const q = "INSERT INTO dbbench.accounts VALUES($1, $2);"
	for i := from; i < to; i++ {
		if _, err := p.db.Exec(q, i, i); err != nil {
			log.Fatalf("failed to insert: %v\n", err)
		}
	}
	return "inserts"
}

func (p *Postgres) selects(from, to int) string {
	const q = "SELECT * FROM dbbench.accounts WHERE id = $1;"
	for i := from; i < to; i++ {
		if _, err := p.db.Exec(q, i); err != nil {
			log.Fatalf("failed to select: %v\n", err)
		}
	}
	return "selects"
}

func (p *Postgres) updates(from, to int) string {
	const q = "UPDATE dbbench.accounts SET balance = $1 WHERE id = $2;"
	for i := from; i < to; i++ {
		if _, err := p.db.Exec(q, i, i); err != nil {
			log.Fatalf("failed to update: %v\n", err)
		}
	}
	return "updates"
}

func (p *Postgres) deletes(from, to int) string {
	const q = "DELETE FROM dbbench.accounts WHERE id = $1"
	for i := from; i < to; i++ {
		if _, err := p.db.Exec(q, i); err != nil {
			log.Fatalf("failed to delete: %v\n", err)
		}
	}
	return "deletes"
}
