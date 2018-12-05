package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sj14/dbbench/benchmark"
	"github.com/sj14/dbbench/databases"
)

var (
	version = "dev version"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var (
		dbType      = flag.String("type", "", "database to use (sqlite|mysql|postgres|cockroach|cassandra)")
		host        = flag.String("host", "localhost", "address of the server")
		port        = flag.Int("port", 0, "port of the server (0 -> db defaults)")
		user        = flag.String("user", "root", "user name to connect with the server")
		pass        = flag.String("pass", "root", "password to connect with the server")
		path        = flag.String("path", "dbbench.sqlite", "database file (sqlite only)")
		conns       = flag.Int("conns", 0, "max. number of open connections")
		iter        = flag.Int("iter", 1000, "how many iterations should be run")
		threads     = flag.Int("threads", 25, "max. number of green threads (iter >= threads > 0")
		sleep       = flag.Duration("sleep", 0, "how long to pause after each single benchmark (valid units: ns, us, ms, s, m, h)")
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

	// we need at least one thread
	if *threads == 0 {
		*threads = 1
	}

	// can't have more threads than iterations
	if *threads > *iter {
		*threads = *iter
	}

	benchmarks := []benchmark.Benchmark{}

	if *scriptname != "" {
		// Benchmark specified script.
		dat, err := ioutil.ReadFile(*scriptname)
		if err != nil {
			log.Fatalf("failed to read file: %v", err)
		}
		buf := bytes.NewBuffer(dat)
		benchmarks = benchmark.ParseScript(buf)
	} else {
		// Use built-in benchmarks.
		benchmarks = bencher.Benchmarks()
	}

	// split benchmark names when "-run 'bench0 bench1 ...'" flag was used
	toRun := strings.Split(*runBench, " ")

	startTotal := time.Now()
	// run benchmarks
	for i, b := range benchmarks {
		// check if we want to run this particular benchmark
		if !contains(toRun, "all") && !contains(toRun, b.Name) {
			continue
		}

		t := template.New(b.Name)
		t, err := t.Parse(b.Stmt)
		if err != nil {
			log.Fatalf("failed to parse template: %v", err)
		}

		start := time.Now()
		if b.Type == benchmark.TypeOnce {
			benchmark.Once(bencher, t)
		} else {
			benchmark.Loop(bencher, t, *iter, *threads)
		}

		elapsed := time.Since(start)
		fmt.Printf("%v:\t%v\t%v\tns/op\n", b.Name, elapsed, elapsed.Nanoseconds()/int64(*iter))

		// Don't sleep after the last benchmark
		if i != len(bencher.Benchmarks())-1 {
			time.Sleep(*sleep)
		}
	}
	fmt.Printf("total: %v\n", time.Since(startTotal))
}

func getImpl(dbType string, host string, port int, user, password, path string, maxOpenConns int) benchmark.Bencher {
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

func contains(options []string, want string) bool {
	for _, o := range options {
		if o == want {
			return true
		}
	}
	return false
}
