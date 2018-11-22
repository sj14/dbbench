package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sj14/dbbench/databases"
)

// Bencher is the interface a benchmark has to impelement.
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
	clean := flag.Bool("clean", false, "only cleanup previous benchmark data, e.g. due to a crash")
	noclean := flag.Bool("noclean", false, "don't cleanup benchmark data")
	// version := flag.Bool("version", false, "print version information") // TODO
	runBench := flag.String("run", "all", "only run the specified benchmark") // TODO

	// subcommands and local flags
	// cassandra := flag.NewFlagSet("cassandra", flag.ExitOnError)
	flag.Parse()

	bencher := getImpl(*db, *host, *port, *user, *pass, *maxOpenConns)

	// try to clean old data and exit when clean flag is set
	if *clean {
		bencher.Cleanup()
		os.Exit(0)
	}

	// setup database
	bencher.Setup()

	// only cleanup benchmark data when noclean flag is not set
	if !*noclean {
		defer bencher.Cleanup()
	}

	start := time.Now()

	// split benchmark names when "-run 'bench0 bench1 ...'" flag was used
	toRun := strings.Split(*runBench, " ")

	for _, r := range toRun {
		benchmark(bencher, r, *iterations, *goroutines)
	}
	fmt.Printf("total: %v\n", time.Since(start))
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

func benchmark(bencher Bencher, runBench string, iterations, goroutines int) {
	wg := &sync.WaitGroup{}
	// w := tabwriter.NewWriter(os.Stdout, 0, 4, 4, '\t', tabwriter.AlignRight)

	for _, b := range bencher.Benchmarks() {
		// check if we want to run this particular benchmark
		if runBench != "all" && b.Name != runBench {
			continue
		}

		wg.Add(goroutines)
		start := time.Now()
		for i := 0; i < goroutines; i++ {
			from := (iterations / goroutines) * i
			to := (iterations / goroutines) * (i + 1)

			go func() {
				defer wg.Done()
				b.Func(from, to)
			}()
		}
		wg.Wait()
		elapsed := time.Since(start)
		fmt.Printf("%v\t%v\t%v\tns/op\n", b.Name, elapsed, elapsed.Nanoseconds()/int64(iterations))
		// fmt.Fprintf(w, "%v\t%v\t%v\tns/op\n", name, elapsed, elapsed.Nanoseconds()/int64(iterations))
	}
	// w.Flush()
}
