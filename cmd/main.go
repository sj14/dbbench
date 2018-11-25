package main

import (
	"bufio"
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
	Setup()
	Cleanup()
	Benchmarks() []databases.Benchmark
	Exec(string)
}

var (
	version = "dev version"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var (
		dbType      = flag.String("type", "", "database to use (sqlite|mariadb|mysql|postgres|cockroach|cassandra|scylla)")
		host        = flag.String("host", "localhost", "address of the server")
		port        = flag.Int("port", 0, "port of the server (0 -> db defaults)")
		user        = flag.String("user", "root", "user name to connect with the server")
		pass        = flag.String("pass", "root", "password to connect with the server")
		path        = flag.String("path", "dbbench.sqlite", "database file (sqlite only)")
		conns       = flag.Int("conns", 0, "max. number of open connections")
		iter        = flag.Int("iter", 1000, "how many iterations should be run")
		threads     = flag.Int("threads", 25, "max. number of green threads")
		nosetup     = flag.Bool("noinit", false, "do not initialize database and tables, e.g. when only running own script")
		clean       = flag.Bool("clean", false, "only cleanup benchmark data, e.g. after a crash")
		noclean     = flag.Bool("noclean", false, "keep benchmark data")
		versionFlag = flag.Bool("version", false, "print version information")
		runBench    = flag.String("run", "all", "only run the specified benchmarks, e.g. \"inserts deletes\"")
		scriptname  = flag.String("script", "", "custom sql file to execute")
	)

	flag.Parse()

	if *versionFlag {
		fmt.Printf("dbbench %v, commit %v, built at %v\n", version, commit, date)
		os.Exit(0)
	}

	bencher := getImpl(*dbType, *host, *port, *user, *pass, *path, *conns)

	// only clean old data when clean flag is set
	if *clean {
		bencher.Cleanup()
		os.Exit(0)
	}

	// setup database
	if !*nosetup {
		bencher.Setup()
	}

	// only cleanup benchmark data when noclean flag is not set
	if !*noclean {
		defer bencher.Cleanup()
	}

	// split benchmark names when "-run 'bench0 bench1 ...'" flag was used
	toRun := strings.Split(*runBench, " ")

	start := time.Now()
	for _, r := range toRun {
		benchmark(bencher, *scriptname, r, *iter, *threads)
	}
	fmt.Printf("total: %v\n", time.Since(start))
}

func getImpl(dbType string, host string, port int, user, password, path string, maxOpenConns int) Bencher {
	switch dbType {
	case "sqlite":
		if maxOpenConns != 0 {
			log.Fatalln("can't use 'conns' with SQLite")
		}
		return databases.NewSQLite(path)
	case "mysql", "mariadb":
		return databases.NewMySQL(host, port, user, password, maxOpenConns)
	case "postgres":
		return databases.NewPostgres(host, port, user, password, maxOpenConns)
	case "cockroach":
		return databases.NewCockroach(host, port, user, password, maxOpenConns)
	case "cassandra", "scylla":
		if maxOpenConns != 0 {
			log.Fatalln("can't use 'conns' with Cassandra or ScyllaDB")
		}
		return databases.NewCassandra(host, port, user, password)
	}

	log.Fatalln("missing or unknown type parameter")
	return nil
}

func benchmark(bencher Bencher, filename, runBench string, iterations, goroutines int) {
	// run specified script
	if filename != "" {
		start := time.Now()
		execScript(bencher, filename, iterations, goroutines)
		elapsed := time.Since(start)
		fmt.Printf("custom script: %v\t%v\tns/op\n", elapsed, elapsed.Nanoseconds()/int64(iterations))
		return
	}

	// run built-in benchmarks
	for _, b := range bencher.Benchmarks() {
		// check if we want to run this particular benchmark
		if runBench != "all" && b.Name != runBench {
			continue
		}

		start := time.Now()
		execBenchmark(b, iterations, goroutines)
		elapsed := time.Since(start)
		fmt.Printf("%v\t%v\t%v\tns/op\n", b.Name, elapsed, elapsed.Nanoseconds()/int64(iterations))
	}
}

func execScript(bencher Bencher, filename string, iterations, goroutines int) {
	lines, err := readLines(filename)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	// store statements in a single line, to execute them at once,
	// otherwise it would cause race conditions with the database and the goroutines
	script := ""
	for _, l := range lines {
		script += l
	}

	wg := &sync.WaitGroup{}
	wg.Add(goroutines)
	defer wg.Wait()

	for i := 0; i < goroutines; i++ {
		from := (iterations / goroutines) * i
		to := (iterations / goroutines) * (i + 1)

		go func() {
			defer wg.Done()
			for i := from; i < to; i++ {
				bencher.Exec(script)
			}
		}()
	}
}

func execBenchmark(b databases.Benchmark, iterations, goroutines int) {
	wg := &sync.WaitGroup{}
	wg.Add(goroutines)
	defer wg.Wait()

	for i := 0; i < goroutines; i++ {
		from := (iterations / goroutines) * i
		to := (iterations / goroutines) * (i + 1)

		go func() {
			defer wg.Done()

			switch b.Type {
			case databases.Single:
				b.Func(1)
			case databases.Loop:
				for i := from; i < to; i++ {
					b.Func(i)
				}
			}
		}()
	}
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}
