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
	if port == 0 {
		port = 9042
	}
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
// TODO: update is not like other db statements balance = balance + balance!
func (c *Cassandra) Benchmarks() []Benchmark {
	return []Benchmark{
		{"inserts", Loop, "INSERT INTO dbbench.accounts (id, balance) VALUES({{.Iter}}, {{.Iter}}) IF NOT EXISTS;"},
		{"updates", Loop, "UPDATE dbbench.accounts SET balance = {{.Iter}} WHERE id = {{.Iter}} IF EXISTS;"},
		{"selects", Loop, "SELECT * FROM dbbench.accounts WHERE id = {{.Iter}};"},
		{"deletes", Loop, "DELETE FROM dbbench.accounts WHERE id = {{.Iter}} IF EXISTS;"},
	}
}

// Setup initializes the database for the benchmark.
func (c *Cassandra) Setup() {
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

// Exec executes the given statement on the database.
func (c *Cassandra) Exec(stmt string) {
	if err := c.session.Query(stmt).Exec(); err != nil {
		log.Fatalf("%v: failed: %v\n", stmt, err)
	}
}
