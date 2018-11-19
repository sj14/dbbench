package cassandra

import (
	"fmt"
	"log"

	"github.com/gocql/gocql"
)

// Cassandra implements the bencher implementation.
type Cassandra struct {
	session *gocql.Session
}

// New returns a new cassandra bencher.
func New(host string, port int, user, password string) *Cassandra {
	dataSourceName := fmt.Sprintf("%v:%v", host, port) // TODO: check how to do with port, user and password

	cluster := gocql.NewCluster(dataSourceName)
	cluster.Keyspace = ""
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("failed to create session: %v\n", err)
	}
	// defer session.Close()

	return &Cassandra{session: session}
}

// Benchmarks returns the individual benchmark functions for the cassandra db.
func (p *Cassandra) Benchmarks() []func(int, int) string {
	return []func(int, int) string{p.inserts, p.updates, p.selects, p.deletes}
}

// Setup initializes the database for the benchmark.
func (p *Cassandra) Setup() {
	// if err := p.session.Query("CREATE DATABASE IF NOT EXISTS dbbench").Exec(); err != nil {
	// 	log.Fatalf("failed to create database: %v\n", err)
	// }
	if err := p.session.Query("CREATE KEYSPACE IF NOT EXISTS dbbench WITH replication = { 'class':'SimpleStrategy', 'replication_factor' : 1 }").Exec(); err != nil {
		log.Fatalf("failed to create keyspace: %v\n", err)
	}
	// TODO: other tests use decemal for balance, cassandra or gocql doesn't to a automatic casting from int to decimal
	if err := p.session.Query("CREATE TABLE IF NOT EXISTS dbbench.accounts (id INT PRIMARY KEY, balance INT);").Exec(); err != nil {
		log.Fatalf("failed to create table: %v\n", err)
	}
	if err := p.session.Query("TRUNCATE dbbench.accounts;").Exec(); err != nil {
		log.Fatalf("failed to truncate table: %v\n", err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (p *Cassandra) Cleanup() {
	if err := p.session.Query("DROP TABLE dbbench.accounts").Exec(); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	if err := p.session.Query("DROP KEYSPACE dbbench").Exec(); err != nil {
		log.Printf("failed to drop database: %v\n", err)
	}
	p.session.Close()
}

func (p *Cassandra) inserts(from, to int) string {
	const q = "INSERT INTO dbbench.accounts (id, balance) VALUES(?, ?);"
	for i := from; i < to; i++ {
		if err := p.session.Query(q, i, i).Exec(); err != nil {
			log.Fatalf("failed to insert: %v\n", err)
		}
	}
	return "inserts"
}

func (p *Cassandra) selects(from, to int) string {
	const q = "SELECT * FROM dbbench.accounts WHERE id = ?;"
	for i := from; i < to; i++ {
		if err := p.session.Query(q, i).Exec(); err != nil {
			log.Fatalf("failed to select: %v\n", err)
		}
	}
	return "selects"
}

func (p *Cassandra) updates(from, to int) string {
	const q = "UPDATE dbbench.accounts SET balance = ? WHERE id = ?;"
	for i := from; i < to; i++ {
		if err := p.session.Query(q, i, i).Exec(); err != nil {
			log.Fatalf("failed to update: %v\n", err)
		}
	}
	return "updates"
}

func (p *Cassandra) deletes(from, to int) string {
	const q = "DELETE FROM dbbench.accounts WHERE id = ?"
	for i := from; i < to; i++ {
		if err := p.session.Query(q, i).Exec(); err != nil {
			log.Fatalf("failed to delete: %v\n", err)
		}
	}
	return "deletes"
}
