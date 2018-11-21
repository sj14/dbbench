package databases

import (
	"database/sql"
	"fmt"
	"log"
)

// Mysql implements the bencher implementation.
type Mysql struct {
	db *sql.DB
}

// New NewMySQL a new mysql bencher.
func NewMySQL(host string, port int, user, password string, maxOpenConns int) *Mysql {
	// username:password@protocol(address)/dbname?param=value
	dataSourceName := fmt.Sprintf("%v:%v@tcp(%v:%v)/dbbench?charset=utf8", user, password, host, port)

	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatalf("failed to open connection: %v\n", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	p := &Mysql{db: db}
	return p
}

// Benchmarks returns the individual benchmark functions for the mysql db.
func (m *Mysql) Benchmarks() []Benchmark {
	return []Benchmark{
		{"inserts", m.inserts},
		{"updates", m.updates},
		{"selects", m.selects},
		{"deletes", m.deletes},
	}
}

// Setup initializes the database for the benchmark.
func (m *Mysql) Setup(...string) {
	if _, err := m.db.Exec("CREATE DATABASE IF NOT EXISTS dbbench"); err != nil {
		log.Fatalf("failed to create database: %v\n", err)
	}
	if _, err := m.db.Exec("CREATE TABLE IF NOT EXISTS dbbench.accounts (id INT PRIMARY KEY, balance DECIMAL);"); err != nil {
		log.Fatalf("failed to create table: %v\n", err)
	}
	if _, err := m.db.Exec("TRUNCATE dbbench.accounts;"); err != nil {
		log.Fatalf("failed to truncate table: %v\n", err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (m *Mysql) Cleanup() {
	if _, err := m.db.Exec("DROP TABLE dbbench.accounts"); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	// if _, err := m.db.Exec("DROP SCHEMA dbbench"); err != nil {
	// 	log.Printf("failed drop schema: %v\n", err)
	// }
	if err := m.db.Close(); err != nil {
		log.Printf("failed to close connection: %v", err)
	}
}

func (m *Mysql) inserts(from, to int) {
	const q = "INSERT INTO dbbench.accounts VALUES(?, ?);"
	for i := from; i < to; i++ {
		if _, err := m.db.Exec(q, i, i); err != nil {
			log.Fatalf("failed to insert: %v\n", err)
		}
	}
}

func (m *Mysql) selects(from, to int) {
	const q = "SELECT * FROM dbbench.accounts WHERE id = ?;"
	for i := from; i < to; i++ {
		if _, err := m.db.Exec(q, i); err != nil {
			log.Fatalf("failed to select: %v\n", err)
		}
	}
}

func (m *Mysql) updates(from, to int) {
	const q = "UPDATE dbbench.accounts SET balance = ? WHERE id = ?;"
	for i := from; i < to; i++ {
		if _, err := m.db.Exec(q, i, i); err != nil {
			log.Fatalf("failed to update: %v\n", err)
		}
	}
}

func (m *Mysql) deletes(from, to int) {
	const q = "DELETE FROM dbbench.accounts WHERE id = ?"
	for i := from; i < to; i++ {
		if _, err := m.db.Exec(q, i); err != nil {
			log.Fatalf("failed to delete: %v\n", err)
		}
	}
}
