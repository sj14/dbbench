package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/sj14/dbbench/cockroach"
	"github.com/sj14/dbbench/postgres"
)

// Bencher is the interface a benchmark has to impelement
type Bencher interface {
	Setup()
	Cleanup()
	Benchmarks() []func(int) string
}

func main() {
	iterations := flag.Int("i", 1000, "how many iterations should be run")
	dbType := flag.String("type", "", "database/driver type to use (postgres|cockroach)")
	host := flag.String("host", "localhsot", "address of the server")
	port := flag.Int("port", 0, "port of the server")
	user := flag.String("user", "root", "user name to connect with the server")
	password := flag.String("password", "root", "password to connect with the server")
	flag.Parse()

	var bencher Bencher
	switch *dbType {
	case "postgres", "pg":
		bencher = postgres.New(*host, *port, *user, *password)
	case "cockroach", "cr":
		bencher = cockroach.New(*host, *port, *user, *password)
	default:
		log.Fatalln("missing type parameter")
	}

	benchmark(bencher, *iterations)
}

func benchmark(impl Bencher, iterations int) {
	defer impl.Cleanup()
	impl.Setup()

	for _, b := range impl.Benchmarks() {
		start := time.Now()
		name := b(iterations)
		fmt.Printf("%v took %v\n", name, time.Now().Sub(start))
	}
}
