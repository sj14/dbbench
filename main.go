package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/sj14/dbbench/cockroach"
	"github.com/sj14/dbbench/postgres"
)

// Bencher is the interface a benchmark has to impelement
type Bencher interface {
	Setup()
	Cleanup()
	Benchmarks() []func(from int, to int) (name string)
}

func main() {
	iterations := flag.Int("i", 1000, "how many iterations should be run")
	dbType := flag.String("type", "", "database/driver type to use (postgres|cockroach)")
	host := flag.String("host", "localhsot", "address of the server")
	port := flag.Int("port", 0, "port of the server")
	user := flag.String("user", "root", "user name to connect with the server")
	password := flag.String("password", "root", "password to connect with the server")
	flag.Parse()

	bencher := getImpl(*dbType, *host, *port, *user, *password)

	bencher.Setup()
	defer bencher.Cleanup()

	benchmark2(bencher, *iterations)
}

func getImpl(dbType string, host string, port int, user, password string) Bencher {
	switch dbType {
	case "postgres", "pg":
		return postgres.New(host, port, user, password)
	case "cockroach", "cr":
		return cockroach.New(host, port, user, password)
	}

	log.Fatalln("missing or unknown type parameter")
	return nil
}

func benchmark2(impl Bencher, iterations int) {
	wg := &sync.WaitGroup{}

	for _, b := range impl.Benchmarks() {
		max := 10
		start := time.Now()
		wg.Add(max)
		var name string
		for i := 0; i < max; i++ {
			from := (iterations / max) * i
			to := (iterations / max) * (i + 1)

			go func() {
				name = b(from, to)
				wg.Done()
			}()
		}
		wg.Wait()
		fmt.Printf("%v took %v\n", name, time.Now().Sub(start))
	}
}
