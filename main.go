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
	flag.Parse()

	var bencher Bencher
	switch *dbType {
	case "postgres", "pg":
		bencher = postgres.New()
	case "cockroach", "cr":
		bencher = cockroach.New()
	default:
		log.Fatalln("missing type parameter")
	}

	benchmark(bencher, *iterations)
}

func benchmark(impl Bencher, iterations int) {
	impl.Setup()

	for _, b := range impl.Benchmarks() {
		start := time.Now()
		name := b(iterations)
		fmt.Printf("%v took %v\n", name, time.Now().Sub(start))
	}

	impl.Cleanup()
}
