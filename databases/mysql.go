package databases

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/sj14/dbbench/benchmark"
)

// Mysql implements the bencher implementation.
type Mysql struct {
	db *sql.DB
}

// NewMySQL returns a new mysql bencher.
func NewMySQL(host string, port int, user, password string, maxOpenConns int) *Mysql {
	if port == 0 {
		port = 3306
	}
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
func (m *Mysql) Benchmarks() []benchmark.Benchmark {
	return []benchmark.Benchmark{
		{"inserts", benchmark.TypeLoop, "INSERT INTO dbbench.simple (id, balance) VALUES( {{.Iter}}, {{call .RandInt63n 9999999999}});"},
		{"selects", benchmark.TypeLoop, "SELECT * FROM dbbench.simple WHERE id = {{.Iter}};"},
		{"updates", benchmark.TypeLoop, "UPDATE dbbench.simple SET balance = {{call .RandInt63n 9999999999}} WHERE id = {{.Iter}};"},
		{"deletes", benchmark.TypeLoop, "DELETE FROM dbbench.simple WHERE id = {{.Iter}};"},
		{"relation_insert0", benchmark.TypeLoop, "INSERT INTO dbbench.relational_one (oid, balance_one) VALUES( {{.Iter}}, {{call .RandInt63n 9999999999}});"},
		{"relation_insert1", benchmark.TypeLoop, "INSERT INTO dbbench.relational_two (relation, balance_two) VALUES( {{.Iter}}, {{call .RandInt63n 9999999999}});"},
		{"relation_select", benchmark.TypeLoop, "SELECT * FROM dbbench.relational_two INNER JOIN dbbench.relational_one ON relational_one.oid = relational_two.relation WHERE relation = {{.Iter}};"},
		{"relation_delete1", benchmark.TypeLoop, "DELETE FROM dbbench.relational_two WHERE relation = {{.Iter}};"},
		{"relation_delete0", benchmark.TypeLoop, "DELETE FROM dbbench.relational_one WHERE oid = {{.Iter}};"},
	}
}

// Setup initializes the database for the benchmark.
func (m *Mysql) Setup() {
	if _, err := m.db.Exec("CREATE DATABASE IF NOT EXISTS dbbench"); err != nil {
		log.Fatalf("failed to create database: %v\n", err)
	}
	if _, err := m.db.Exec("CREATE TABLE IF NOT EXISTS dbbench.simple (id INT PRIMARY KEY, balance DECIMAL);"); err != nil {
		log.Fatalf("failed to create table: %v\n", err)
	}
	if _, err := m.db.Exec("CREATE TABLE IF NOT EXISTS dbbench.relational_one (oid INT PRIMARY KEY, balance_one DECIMAL);"); err != nil {
		log.Fatalf("failed to create table relational_one: %v\n", err)
	}
	if _, err := m.db.Exec("CREATE TABLE IF NOT EXISTS dbbench.relational_two (balance_two DECIMAL, relation INT, FOREIGN KEY(relation) REFERENCES relational_one(oid));"); err != nil {
		log.Fatalf("failed to create table relational_two: %v\n", err)
	}
	if _, err := m.db.Exec("TRUNCATE dbbench.simple;"); err != nil {
		log.Fatalf("failed to truncate table: %v\n", err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (m *Mysql) Cleanup() {
	if _, err := m.db.Exec("DROP TABLE dbbench.simple"); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	// When the database will be dropped here,
	// the tool is not able recreate it during setup.
	//
	// if _, err := m.db.Exec("DROP DATABASE dbbench"); err != nil {
	// 	log.Printf("failed drop schema: %v\n", err)
	// }
	if _, err := m.db.Exec("DROP TABLE dbbench.relational_two"); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	if _, err := m.db.Exec("DROP TABLE dbbench.relational_one"); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	if err := m.db.Close(); err != nil {
		log.Printf("failed to close connection: %v", err)
	}
}

// Exec executes the given statement on the database.
func (m *Mysql) Exec(stmt string) {
	_, err := m.db.Exec(stmt)
	if err != nil {
		log.Printf("%v failed: %v", stmt, err)
	}
}
