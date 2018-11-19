package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql" // mysql db driver
	_ "github.com/lib/pq"              // pq is the postgres/cockroach db driver
	"github.com/sj14/dbbench/cassandra"
	"github.com/sj14/dbbench/cockroach"
	"github.com/sj14/dbbench/mysql"
	"github.com/sj14/dbbench/postgres"
)

// Bencher is the interface a benchmark has to impelement
type Bencher interface {
	Setup()
	Cleanup()
	Benchmarks() []func(from int, to int) (name string)
}

func main() {
	// TODO: add database name flag (especially for mysql)
	iterations := flag.Int("i", 1000, "how many iterations should be run")
	db := flag.String("db", "", "database/driver to use (mysql|postgres|cockroach)")
	goroutines := flag.Int("threads", 10, "how many green threads (goroutines) to use")
	host := flag.String("host", "localhsot", "address of the server")
	port := flag.Int("port", 0, "port of the server")
	user := flag.String("user", "root", "user name to connect with the server")
	password := flag.String("password", "root", "password to connect with the server")
	flag.Parse()

	bencher := getImpl(*db, *host, *port, *user, *password)

	bencher.Setup()
	defer bencher.Cleanup()

	benchmark(bencher, *iterations, *goroutines)
}

func getImpl(dbType string, host string, port int, user, password string) Bencher {
	switch dbType {
	case "mysql", "mariadb":
		return mysql.New(host, port, user, password)
	case "postgres", "pg":
		return postgres.New(host, port, user, password)
	case "cockroach", "cr":
		return cockroach.New(host, port, user, password)
	case "cassandra":
		return cassandra.New(host, port, user, password)
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
