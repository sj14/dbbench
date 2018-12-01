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

// Benchmarks returns the individual benchmark statements for the postgres db.
func (p *Postgres) Benchmarks() []Benchmark {
	return []Benchmark{
		{"inserts", Loop, "INSERT INTO dbbench.dbbench_simple (id, balance) VALUES( {{.Iter}}, {{call .RandInt63}});"},
		{"selects", Loop, "SELECT * FROM dbbench.dbbench_simple WHERE id = {{.Iter}};"},
		{"updates", Loop, "UPDATE dbbench.dbbench_simple SET balance = {{call .RandInt63}} WHERE id = {{.Iter}};"},
		{"deletes", Loop, "DELETE FROM dbbench.dbbench_simple WHERE id = {{.Iter}};"},
		{"relation_insert0", Loop, "INSERT INTO dbbench.dbbench_relational_one (oid, balance_one) VALUES( {{.Iter}}, {{call .RandInt63}});"},
		{"relation_insert1", Loop, "INSERT INTO dbbench.dbbench_relational_two (relation, balance_two) VALUES( {{.Iter}}, {{call .RandInt63}});"},
	}
}

// Setup initializes the database for the benchmark.
func (p *Postgres) Setup() {
	if _, err := p.db.Exec("CREATE SCHEMA IF NOT EXISTS dbbench"); err != nil {
		log.Fatalf("failed to create schema: %v\n", err)
	}
	if _, err := p.db.Exec("CREATE TABLE IF NOT EXISTS dbbench.dbbench_simple (id INT PRIMARY KEY, balance DECIMAL);"); err != nil {
		log.Fatalf("failed to create table: %v\n", err)
	}
	if _, err := p.db.Exec("CREATE TABLE IF NOT EXISTS dbbench.dbbench_relational_one (oid INT PRIMARY KEY, balance_one DECIMAL);"); err != nil {
		log.Fatalf("failed to create table dbbench_relational_one: %v\n", err)
	}
	if _, err := p.db.Exec("CREATE TABLE IF NOT EXISTS dbbench.dbbench_relational_two (balance_two DECIMAL, relation INT, FOREIGN KEY(relation) REFERENCES dbbench.dbbench_relational_one(oid));"); err != nil {
		log.Fatalf("failed to create table dbbench_relational_two: %v\n", err)
	}
	if _, err := p.db.Exec("TRUNCATE dbbench.dbbench_simple;"); err != nil {
		log.Fatalf("failed to truncate table: %v\n", err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (p *Postgres) Cleanup() {
	if _, err := p.db.Exec("DROP TABLE dbbench.dbbench_simple"); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	if _, err := p.db.Exec("DROP TABLE dbbench.dbbench_relational_two"); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	if _, err := p.db.Exec("DROP TABLE dbbench.dbbench_relational_one"); err != nil {
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
