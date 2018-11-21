package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sj14/dbbench/databases"
)

// Bencher is the interface a benchmark has to impelement
type Bencher interface {
	Setup(...string)
	Cleanup()
	Benchmarks() []databases.Benchmark
}

func main() {
	// TODO: add database name flag (especially for mysql)
	// global flags
	db := flag.String("db", "", "database to use (sqlite|mariadb|mysql|postgres|cockroach|cassandra|scylla)")
	host := flag.String("host", "localhost", "address of the server")
	port := flag.Int("port", 0, "port of the server")
	user := flag.String("user", "root", "user name to connect with the server")
	pass := flag.String("pass", "root", "password to connect with the server")
	iterations := flag.Int("iter", 1000, "how many iterations should be run")
	goroutines := flag.Int("threads", 25, "max. number of green threads")
	maxOpenConns := flag.Int("conns", 0, "max. number of open connections")
	clean := flag.Bool("clean", false, "only cleanup previous benchmark data, e.g. due to a crash (no benchmark will run)")
	// runBench := flag.String("run", "all", "only run the specified benchmark") // TODO

	// subcommands and local flags
	// cassandra := flag.NewFlagSet("cassandra", flag.ExitOnError)
	flag.Parse()

	bencher := getImpl(*db, *host, *port, *user, *pass, *maxOpenConns)

	if *clean {
		// Clean flag. Try to clean old data and exit.
		bencher.Cleanup()
		os.Exit(0)
	}

	bencher.Setup()
	defer bencher.Cleanup()

	benchmark(bencher, *iterations, *goroutines)
}

func getImpl(dbType string, host string, port int, user, password string, maxOpenConns int) Bencher {
	switch dbType {
	case "sqlite":
		if maxOpenConns != 0 {
			log.Fatalln("can't use flag 'conns' for SQLite")
		}
		return databases.NewSQLite()
	case "mysql", "mariadb":
		return databases.NewMySQL(host, port, user, password, maxOpenConns)
	case "postgres":
		return databases.NewPostgres(host, port, user, password, maxOpenConns)
	case "cockroach":
		return databases.NewCockroach(host, port, user, password, maxOpenConns)
	case "cassandra", "scylla":
		if maxOpenConns != 0 {
			log.Fatalln("can't use flag 'conns' for Cassandra or ScyllaDB")
		}
		return databases.NewCassandra(host, port, user, password)
	}

	log.Fatalln("missing or unknown type parameter")
	return nil
}

func benchmark(bencher Bencher, iterations, goroutines int) {
	wg := &sync.WaitGroup{}

	for _, b := range bencher.Benchmarks() {
		wg.Add(goroutines)
		var name string
		start := time.Now()
		for i := 0; i < goroutines; i++ {
			from := (iterations / goroutines) * i
			to := (iterations / goroutines) * (i + 1)

			go func() {
				name = b.Name
				b.Func(from, to)
				wg.Done()
			}()
		}
		wg.Wait()
		fmt.Printf("%v took %v\n", name, time.Now().Sub(start))
	}
}
