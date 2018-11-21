package databases

import (
	"fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
)

// Cassandra implements the bencher implementation.
type Cassandra struct {
	session *gocql.Session
}

// NewCassandra returns a new cassandra bencher.
func NewCassandra(host string, port int, user, password string) *Cassandra {
	dataSourceName := fmt.Sprintf("%v:%v", host, port) // TODO: check how to do with port, user and password

	cluster := gocql.NewCluster(dataSourceName)
	cluster.Keyspace = ""
	cluster.Timeout = 5 * time.Minute
	cluster.Consistency = gocql.Quorum
	// TOOD: as flags
	// Any         Consistency = 0x00
	// One         Consistency = 0x01
	// Two         Consistency = 0x02
	// Three       Consistency = 0x03
	// Quorum      Consistency = 0x04
	// All         Consistency = 0x05
	// LocalQuorum Consistency = 0x06
	// EachQuorum  Consistency = 0x07
	// LocalOne    Consistency = 0x0A
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("failed to create session: %v\n", err)
	}

	return &Cassandra{session: session}
}

// Benchmarks returns the individual benchmark functions for the cassandra db.
func (c *Cassandra) Benchmarks() []Benchmark {
	return []Benchmark{
		{"inserts", c.inserts},
		{"updates", c.updates},
		{"selects", c.selects},
		{"deletes", c.deletes},
	}
}

// Setup initializes the database for the benchmark.
func (c *Cassandra) Setup(...string) {
	// TODO: flags for class and replication factor
	if err := c.session.Query("CREATE KEYSPACE IF NOT EXISTS dbbench WITH replication = { 'class':'SimpleStrategy', 'replication_factor' : 1 }").Exec(); err != nil {
		log.Fatalf("failed to create keyspace: %v\n", err)
	}
	// TODO: other tests use decimal for balance, cassandra or gocql doesn't do an automatic casting from int to decimal
	if err := c.session.Query("CREATE TABLE IF NOT EXISTS dbbench.accounts (id INT PRIMARY KEY, balance INT);").Exec(); err != nil {
		log.Fatalf("failed to create table: %v\n", err)
	}
	if err := c.session.Query("TRUNCATE dbbench.accounts;").Exec(); err != nil {
		log.Fatalf("failed to truncate table: %v\n", err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (c *Cassandra) Cleanup() {
	if err := c.session.Query("DROP TABLE dbbench.accounts").Exec(); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	if err := c.session.Query("DROP KEYSPACE dbbench").Exec(); err != nil {
		log.Printf("failed to drop database: %v\n", err)
	}
	c.session.Close()
}

func (c *Cassandra) inserts(from, to int) {
	const q = "INSERT INTO dbbench.accounts (id, balance) VALUES(?, ?);"
	for i := from; i < to; i++ {
		if err := c.session.Query(q, i, i).Exec(); err != nil {
			log.Fatalf("failed to insert: %v\n", err)
		}
	}
}

func (c *Cassandra) selects(from, to int) {
	const q = "SELECT * FROM dbbench.accounts WHERE id = ?;"
	for i := from; i < to; i++ {
		if err := c.session.Query(q, i).Exec(); err != nil {
			log.Fatalf("failed to select: %v\n", err)
		}
	}
}

func (c *Cassandra) updates(from, to int) {
	const q = "UPDATE dbbench.accounts SET balance = ? WHERE id = ?;"
	for i := from; i < to; i++ {
		if err := c.session.Query(q, i, i).Exec(); err != nil {
			log.Fatalf("failed to update: %v\n", err)
		}
	}
}

func (c *Cassandra) deletes(from, to int) {
	const q = "DELETE FROM dbbench.accounts WHERE id = ?"
	for i := from; i < to; i++ {
		if err := c.session.Query(q, i).Exec(); err != nil {
			log.Fatalf("failed to delete: %v\n", err)
		}
	}
}
