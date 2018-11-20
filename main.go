package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql" // mysql db driver
	_ "github.com/lib/pq"              // pq is the postgres/cockroach db driver
	"github.com/sj14/dbbench/databases"
)

// Bencher is the interface a benchmark has to impelement
type Bencher interface {
	Setup(...string)
	Cleanup()
	Benchmarks() []func(from int, to int) (name string)
}

func main() {
	// TODO: add database name flag (especially for mysql)
	// global flags
	db := flag.String("db", "", "database/driver to use (mysql|postgres|cockroach)")
	host := flag.String("host", "localhost", "address of the server")
	port := flag.Int("port", 0, "port of the server")
	user := flag.String("user", "root", "user name to connect with the server")
	pass := flag.String("pass", "root", "password to connect with the server")
	iterations := flag.Int("iter", 1000, "how many iterations should be run")
	goroutines := flag.Int("threads", 25, "max. number of green threads (goroutines)")
	maxOpenConns := flag.Int("conns", 0, "max. number of open connections")

	// subcommands and local flags
	// cassandra := flag.NewFlagSet("cassandra", flag.ExitOnError)

	flag.Parse()

	bencher := getImpl(*db, *host, *port, *user, *pass, *maxOpenConns)

	bencher.Setup()
	defer bencher.Cleanup()

	benchmark(bencher, *iterations, *goroutines)
}

func getImpl(dbType string, host string, port int, user, password string, maxOpenConns int) Bencher {
	switch dbType {
	case "mysql", "mariadb":
		return databases.NewMySQL(host, port, user, password, maxOpenConns)
	case "postgres", "pg":
		return databases.NewPostgres(host, port, user, password, maxOpenConns)
	case "cockroach", "cr":
		return databases.NewCockroach(host, port, user, password, maxOpenConns)
	case "cassandra", "scylla":
		if maxOpenConns != 0 {
			log.Fatalln("can't use flag 'conns' for cassandra or scylla")
		}
		return databases.NewCassandra(host, port, user, password)
	}

	log.Fatalln("missing or unknown type parameter")
	return nil
}

func benchmark(impl Bencher, iterations, goroutines int) {
	wg := &sync.WaitGroup{}

	for _, b := range impl.Benchmarks() {
		wg.Add(goroutines)
		var name string
		start := time.Now()
		for i := 0; i < goroutines; i++ {
			from := (iterations / goroutines) * i
			to := (iterations / goroutines) * (i + 1)

			go func() {
				name = b(from, to)
				wg.Done()
			}()
		}
		wg.Wait()
		fmt.Printf("%v took %v\n", name, time.Now().Sub(start))
	}
}
