package cockroach

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq" // pq is the postgres/cockroach db driver
)

// Cockroach implements the bencher implementation.
type Cockroach struct {
	// db []*sql.DB
	db *sql.DB
	// inChan chan int
}

// New returns a new cockroach bencher.
func New(host string, port int, user, password string) *Cockroach {
	dataSourceName := fmt.Sprintf("host=%v port=%v user='%v' password='%v' sslmode=disable", host, port, user, password)

	c := &Cockroach{}

	// for i := 1; i < 5; i++ {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatalf("failed to open connection: %v\n", err)
	}

	// c.db = append(c.db, db)
	// }

	c.db = db
	return c
}

// Benchmarks returns the individual benchmark functions for the cockroach db.
func (p *Cockroach) Benchmarks() []func(*sync.WaitGroup, int, int) string {
	return []func(*sync.WaitGroup, int, int) string{p.inserts, p.updates, p.selects, p.deletes}
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
	p.db.Close()
}

func (p *Cockroach) inserts(wg *sync.WaitGroup, from, to int) string {
	const q = "INSERT INTO dbbench.accounts VALUES($1, $2);"
	for i := from; i < to; i++ {
		if _, err := p.db.Exec(q, i, i); err != nil {
			log.Fatalf("failed to insert: %v\n", err)
		}
	}
	wg.Done()
	return "inserts"
}

func (p *Cockroach) selects(wg *sync.WaitGroup, from, to int) string {
	const q = "SELECT * FROM dbbench.accounts WHERE id = $1;"
	for i := from; i < to; i++ {
		if _, err := p.db.Exec(q, i); err != nil {
			log.Fatalf("failed to select: %v\n", err)
		}
	}
	wg.Done()
	return "selects"
}

func (p *Cockroach) updates(wg *sync.WaitGroup, from, to int) string {
	const q = "UPDATE dbbench.accounts SET balance = $1 WHERE id = $2;"
	for i := from; i < to; i++ {
		if _, err := p.db.Exec(q, i, i); err != nil {
			log.Fatalf("failed to update: %v\n", err)
		}
	}
	wg.Done()
	return "updates"
}

func (p *Cockroach) deletes(wg *sync.WaitGroup, from, to int) string {
	const q = "DELETE FROM dbbench.accounts WHERE id = $1"
	for i := from; i < to; i++ {
		if _, err := p.db.Exec(q, i); err != nil {
			log.Fatalf("failed to delete: %v\n", err)
		}
	}
	wg.Done()
	return "deletes"
}
