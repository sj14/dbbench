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
		{"inserts", Loop, "INSERT INTO dbbench.dbbench_simple (id, balance) VALUES( {{.Iter}}, {{call .RandInt63}});"},
		{"selects", Loop, "SELECT * FROM dbbench.dbbench_simple WHERE id = {{.Iter}};"},
		{"updates", Loop, "UPDATE dbbench.dbbench_simple SET balance = {{call .RandInt63}} WHERE id = {{.Iter}};"},
		{"deletes", Loop, "DELETE FROM dbbench.dbbench_simple WHERE id = {{.Iter}};"},
	}
}

// Setup initializes the database for the benchmark.
func (p *Cockroach) Setup() {
	if _, err := p.db.Exec("CREATE DATABASE IF NOT EXISTS dbbench"); err != nil {
		log.Fatalf("failed to create database: %v\n", err)
	}
	if _, err := p.db.Exec("CREATE TABLE IF NOT EXISTS dbbench.dbbench_simple (id INT PRIMARY KEY, balance DECIMAL);"); err != nil {
		log.Fatalf("failed to create table: %v\n", err)
	}
	if _, err := p.db.Exec("TRUNCATE dbbench.dbbench_simple;"); err != nil {
		log.Fatalf("failed to truncate table: %v\n", err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (p *Cockroach) Cleanup() {
	if _, err := p.db.Exec("DROP TABLE dbbench.dbbench_simple"); err != nil {
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
	_, err := p.db.Exec(stmt)
	if err != nil {
		log.Printf("%v failed: %v", stmt, err)
	}
}

func (p *Cockroach) inserts(i int) {
	const q = "INSERT INTO dbbench.dbbench_simple VALUES($1, $2);"
	result, err := p.db.Exec(q, i, i)
	mustExec(result, err, "insert")
}

func (p *Cockroach) selects(i int) {
	const q = "SELECT * FROM dbbench.dbbench_simple WHERE id = $1;"
	result, err := p.db.Exec(q, i)
	mustExec(result, err, "select")

}

func (p *Cockroach) updates(i int) {
	const q = "UPDATE dbbench.dbbench_simple SET balance = $1 WHERE id = $2;"
	result, err := p.db.Exec(q, i, i)
	mustExec(result, err, "update")
}

func (p *Cockroach) deletes(i int) {
	const q = "DELETE FROM dbbench.dbbench_simple WHERE id = $1"
	result, err := p.db.Exec(q, i)
	mustExec(result, err, "delete")
}
