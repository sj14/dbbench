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
		{"inserts", Loop, "INSERT INTO dbbench.simple (id, balance) VALUES( {{.Iter}}, {{call .RandInt63}});"},
		{"selects", Loop, "SELECT * FROM dbbench.simple WHERE id = {{.Iter}};"},
		{"updates", Loop, "UPDATE dbbench.simple SET balance = {{call .RandInt63}} WHERE id = {{.Iter}};"},
		{"deletes", Loop, "DELETE FROM dbbench.simple WHERE id = {{.Iter}};"},
		{"relation_insert0", Loop, "INSERT INTO dbbench.relational_one (oid, balance_one) VALUES( {{.Iter}}, {{call .RandInt63}});"},
		{"relation_insert1", Loop, "INSERT INTO dbbench.relational_two (relation, balance_two) VALUES( {{.Iter}}, {{call .RandInt63}});"},
		{"relation_select", Loop, "SELECT * FROM dbbench.relational_two INNER JOIN dbbench.relational_one ON relational_one.oid = relational_two.relation WHERE relation = {{.Iter}};"},
		{"relation_delete1", Loop, "DELETE FROM dbbench.relational_two WHERE relation = {{.Iter}};"},
		{"relation_delete0", Loop, "DELETE FROM dbbench.relational_one WHERE oid = {{.Iter}};"},
	}
}

// Setup initializes the database for the benchmark.
func (p *Cockroach) Setup() {
	if _, err := p.db.Exec("CREATE DATABASE IF NOT EXISTS dbbench"); err != nil {
		log.Fatalf("failed to create database: %v\n", err)
	}
	if _, err := p.db.Exec("CREATE TABLE IF NOT EXISTS dbbench.simple (id INT PRIMARY KEY, balance DECIMAL);"); err != nil {
		log.Fatalf("failed to create table: %v\n", err)
	}
	if _, err := p.db.Exec("CREATE TABLE IF NOT EXISTS dbbench.relational_one (oid INT PRIMARY KEY, balance_one DECIMAL);"); err != nil {
		log.Fatalf("failed to create table relational_one: %v\n", err)
	}
	if _, err := p.db.Exec("CREATE TABLE IF NOT EXISTS dbbench.relational_two (balance_two DECIMAL, relation INT, FOREIGN KEY(relation) REFERENCES dbbench.relational_one(oid));"); err != nil {
		log.Fatalf("failed to create table relational_two: %v\n", err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (p *Cockroach) Cleanup() {
	if _, err := p.db.Exec("DROP TABLE dbbench.simple"); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	if _, err := p.db.Exec("DROP TABLE dbbench.relational_two"); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	if _, err := p.db.Exec("DROP TABLE dbbench.relational_one"); err != nil {
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
	const q = "INSERT INTO dbbench.simple VALUES($1, $2);"
	result, err := p.db.Exec(q, i, i)
	mustExec(result, err, "insert")
}

func (p *Cockroach) selects(i int) {
	const q = "SELECT * FROM dbbench.simple WHERE id = $1;"
	result, err := p.db.Exec(q, i)
	mustExec(result, err, "select")

}

func (p *Cockroach) updates(i int) {
	const q = "UPDATE dbbench.simple SET balance = $1 WHERE id = $2;"
	result, err := p.db.Exec(q, i, i)
	mustExec(result, err, "update")
}

func (p *Cockroach) deletes(i int) {
	const q = "DELETE FROM dbbench.simple WHERE id = $1"
	result, err := p.db.Exec(q, i)
	mustExec(result, err, "delete")
}
