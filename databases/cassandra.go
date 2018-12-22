package databases

import (
	"fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
	"github.com/sj14/dbbench/benchmark"
)

// Cassandra implements the bencher interface.
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
	// TODO: as flags?
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
func (c *Cassandra) Benchmarks() []benchmark.Benchmark {
	return []benchmark.Benchmark{
		{Name: "inserts", Type: benchmark.TypeLoop, Stmt: "INSERT INTO dbbench.dbbench_simple (id, balance) VALUES({{.Iter}}, {{call .RandInt63}}) IF NOT EXISTS;"},
		{Name: "selects", Type: benchmark.TypeLoop, Stmt: "SELECT * FROM dbbench.dbbench_simple WHERE id = {{.Iter}};"},
		{Name: "updates", Type: benchmark.TypeLoop, Stmt: "UPDATE dbbench.dbbench_simple SET balance = {{call .RandInt63}} WHERE id = {{.Iter}} IF EXISTS;"},
		{Name: "deletes", Type: benchmark.TypeLoop, Stmt: "DELETE FROM dbbench.dbbench_simple WHERE id = {{.Iter}} IF EXISTS;"},
	}
}

// Setup initializes the database for the benchmark.
func (c *Cassandra) Setup() {
	// TODO: flags for class and replication factor
	if err := c.session.Query("CREATE KEYSPACE IF NOT EXISTS dbbench WITH replication = { 'class':'SimpleStrategy', 'replication_factor' : 1 }").Exec(); err != nil {
		log.Fatalf("failed to create keyspace: %v\n", err)
	}
	if err := c.session.Query("CREATE TABLE IF NOT EXISTS dbbench.dbbench_simple (id INT PRIMARY KEY, balance DECIMAL);").Exec(); err != nil {
		log.Fatalf("failed to create table: %v\n", err)
	}
	if err := c.session.Query("TRUNCATE dbbench.dbbench_simple;").Exec(); err != nil {
		log.Fatalf("failed to truncate table: %v\n", err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (c *Cassandra) Cleanup() {
	if err := c.session.Query("DROP TABLE dbbench.dbbench_simple").Exec(); err != nil {
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
