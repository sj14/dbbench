package databases

import (
	"database/sql"
	"fmt"
	"log"
)

// Postgres implements the bencher implementation.
type Postgres struct {
	db *sql.DB
}

// NewPostgres returns a new postgres bencher.
func NewPostgres(host string, port int, user, password string, maxOpenConns int) *Postgres {
	if port == 0 {
		port = 5432
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

	p := &Postgres{db: db}
	return p
}

// Benchmarks returns the individual benchmark functions for the postgres db.
func (p *Postgres) Benchmarks() []Benchmark {
	return []Benchmark{
		{"inserts", Loop, p.inserts},
		{"updates", Loop, p.updates},
		{"selects", Loop, p.selects},
		{"deletes", Loop, p.deletes},
	}
}

// Setup initializes the database for the benchmark.
func (p *Postgres) Setup() {
	if _, err := p.db.Exec("CREATE SCHEMA IF NOT EXISTS dbbench"); err != nil {
		log.Fatalf("failed to create schema: %v\n", err)
	}
	if _, err := p.db.Exec("CREATE TABLE IF NOT EXISTS dbbench.accounts (id INT PRIMARY KEY, balance DECIMAL);"); err != nil {
		log.Fatalf("failed to create table: %v\n", err)
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

// Exec executes the given statement on the database.
func (p *Postgres) Exec(stmt string) {
	_, err := p.db.Exec(stmt)
	if err != nil {
		log.Printf("%v failed: %v", stmt, err)
	}
}

func (p *Postgres) inserts(i int) {
	const q = "INSERT INTO dbbench.accounts VALUES($1, $2);"
	result, err := p.db.Exec(q, i, i)
	mustExec(result, err, "insert")
}

func (p *Postgres) selects(i int) {
	const q = "SELECT * FROM dbbench.accounts WHERE id = $1;"
	result, err := p.db.Exec(q, i)
	mustExec(result, err, "select")
}

func (p *Postgres) updates(i int) {
	const q = "UPDATE dbbench.accounts SET balance = $1 WHERE id = $2;"
	result, err := p.db.Exec(q, i, i)
	mustExec(result, err, "update")
}

func (p *Postgres) deletes(i int) {
	const q = "DELETE FROM dbbench.accounts WHERE id = $1"
	result, err := p.db.Exec(q, i)
	mustExec(result, err, "delete")
}
