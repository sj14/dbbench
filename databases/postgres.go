package databases

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/sj14/dbbench/benchmark"
)

// Postgres implements the bencher interface.
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
func (p *Postgres) Benchmarks() []benchmark.Benchmark {
	return []benchmark.Benchmark{
		{Name: "inserts", Type: benchmark.TypeLoop, Stmt: "INSERT INTO dbbench.simple (id, balance) VALUES( {{.Iter}}, {{call .RandInt64}});"},
		{Name: "selects", Type: benchmark.TypeLoop, Stmt: "SELECT * FROM dbbench.simple WHERE id = {{.Iter}};"},
		{Name: "updates", Type: benchmark.TypeLoop, Stmt: "UPDATE dbbench.simple SET balance = {{call .RandInt64}} WHERE id = {{.Iter}};"},
		{Name: "deletes", Type: benchmark.TypeLoop, Stmt: "DELETE FROM dbbench.simple WHERE id = {{.Iter}};"},
		// {"relation_insert0", benchmark.TypeLoop, "INSERT INTO dbbench.relational_one (oid, balance_one) VALUES( {{.Iter}}, {{call .RandInt64}});"},
		// {"relation_insert1", benchmark.TypeLoop, "INSERT INTO dbbench.relational_two (relation, balance_two) VALUES( {{.Iter}}, {{call .RandInt64}});"},
		// {"relation_select", benchmark.TypeLoop, "SELECT * FROM dbbench.relational_two INNER JOIN dbbench.relational_one ON relational_one.oid = relational_two.relation WHERE relation = {{.Iter}};"},
		// {"relation_delete1", benchmark.TypeLoop, "DELETE FROM dbbench.relational_two WHERE relation = {{.Iter}};"},
		// {"relation_delete0", benchmark.TypeLoop, "DELETE FROM dbbench.relational_one WHERE oid = {{.Iter}};"},
	}
}

// Setup initializes the database for the benchmark.
func (p *Postgres) Setup() {
	if _, err := p.db.Exec("CREATE SCHEMA IF NOT EXISTS dbbench"); err != nil {
		log.Fatalf("failed to create schema: %v\n", err)
	}
	if _, err := p.db.Exec("CREATE TABLE IF NOT EXISTS dbbench.simple (id INT PRIMARY KEY, balance DECIMAL);"); err != nil {
		log.Fatalf("failed to create table: %v\n", err)
	}
	if _, err := p.db.Exec("CREATE TABLE IF NOT EXISTS dbbench.relational_one (oid INT PRIMARY KEY, balance_one DECIMAL);"); err != nil {
		log.Fatalf("failed to create table relational_one: %v\n", err)
	}
	if _, err := p.db.Exec("CREATE TABLE IF NOT EXISTS dbbench.relational_two (balance_two DECIMAL, relation INT PRIMARY KEY, FOREIGN KEY(relation) REFERENCES dbbench.relational_one(oid));"); err != nil {
		log.Fatalf("failed to create table relational_two: %v\n", err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (p *Postgres) Cleanup() {
	if _, err := p.db.Exec("DROP TABLE dbbench.simple"); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	if _, err := p.db.Exec("DROP TABLE dbbench.relational_two"); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	if _, err := p.db.Exec("DROP TABLE dbbench.relational_one"); err != nil {
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
