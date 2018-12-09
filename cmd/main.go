package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sj14/dbbench/benchmark"
	"github.com/sj14/dbbench/databases"
	"github.com/spf13/pflag"
)

var (
	version = "dev version"
	commit  = "none"
	date    = time.Now().String()
)

func main() {
	var (
		// Default set of flags, available for all subcommands (benchmark options).
		defaults    = pflag.NewFlagSet("defaults", pflag.ExitOnError)
		iter        = defaults.Int("iter", 1000, "how many iterations should be run")
		threads     = defaults.Int("threads", 25, "max. number of green threads (iter >= threads > 0)")
		sleep       = defaults.Duration("sleep", 0, "how long to pause after each single benchmark (valid units: ns, us, ms, s, m, h)")
		nosetup     = defaults.Bool("noinit", false, "do not initialize database and tables, e.g. when only running own script")
		clean       = defaults.Bool("clean", false, "only cleanup benchmark data, e.g. after a crash")
		noclean     = defaults.Bool("noclean", false, "keep benchmark data")
		versionFlag = defaults.Bool("version", false, "print version information")
		runBench    = defaults.String("run", "all", "only run the specified benchmarks, e.g. \"inserts deletes\"")
		scriptname  = defaults.String("script", "", "custom sql file to execute")

		// Connection flags, applicable for most databases (not sqlite).
		conn = pflag.NewFlagSet("conn", pflag.ExitOnError)
		host = conn.String("host", "localhost", "address of the server")
		port = conn.Int("port", 0, "port of the server (0 -> db defaults)")
		user = conn.String("user", "root", "user name to connect with the server")
		pass = conn.String("pass", "root", "password to connect with the server")

		// Max. connections, applicable for most databases (not cassandra, sqlite).
		maxconnsSet = pflag.NewFlagSet("conns", pflag.ExitOnError)
		maxconns    = maxconnsSet.Int("conns", 0, "max. number of open connections")

		// Flag sets for each database. DB specific flags are set in the switch statement below.
		postgres  = pflag.NewFlagSet("postgres", pflag.ExitOnError)
		mysql     = pflag.NewFlagSet("mysql", pflag.ExitOnError)
		sqlite    = pflag.NewFlagSet("sqlite", pflag.ExitOnError)
		cassandra = pflag.NewFlagSet("cassandra", pflag.ExitOnError)
		mssql     = pflag.NewFlagSet("mssql", pflag.ExitOnError)
		cockroach = pflag.NewFlagSet("cockroach", pflag.ExitOnError)
	)

	var bencher benchmark.Bencher
	switch os.Args[1] {
	case "postgres":
		postgres.AddFlagSet(defaults)
		postgres.AddFlagSet(conn)
		postgres.AddFlagSet(maxconnsSet)
		postgres.Parse(os.Args[2:])
		bencher = databases.NewPostgres(*host, *port, *user, *pass, *maxconns)
	case "cockroach":
		cockroach.AddFlagSet(defaults)
		cockroach.AddFlagSet(conn)
		cockroach.AddFlagSet(maxconnsSet)
		cockroach.Parse(os.Args[2:])
		bencher = databases.NewCockroach(*host, *port, *user, *pass, *maxconns)
	case "cassandra", "scylla":
		cassandra.AddFlagSet(defaults)
		cassandra.AddFlagSet(conn)
		cockroach.Parse(os.Args[2:])
		bencher = databases.NewCassandra(*host, *port, *user, *pass)
	case "mysql", "mariadb":
		mysql.AddFlagSet(defaults)
		mysql.AddFlagSet(conn)
		mysql.AddFlagSet(maxconnsSet)
		mysql.Parse(os.Args[2:])
		bencher = databases.NewMySQL(*host, *port, *user, *pass, *maxconns)
	case "mssql":
		mssql.AddFlagSet(defaults)
		mssql.AddFlagSet(conn)
		mssql.AddFlagSet(maxconnsSet)
		mssql.Parse(os.Args[2:])
		bencher = databases.NewMSSQL(*host, *port, *user, *pass, *maxconns)
	case "sqlite":
		sqlite.AddFlagSet(defaults)
		path := sqlite.String("path", "dbbench.sqlite", "database file (sqlite only)")
		sqlite.Parse(os.Args[2:])
		bencher = databases.NewSQLite(*path)
	default:
		defaults.Usage = func() {
			fmt.Fprintf(os.Stderr, "Available subcommands:\n\tcassandra|cockroach|mssql|mysql|postgres|sqlite\n")
			fmt.Fprintf(os.Stderr, "\tUse 'subcommand --help' for all flags of the specified command.\n")
			fmt.Fprintf(os.Stderr, "Generic flags for all subcommands:\n")
			defaults.PrintDefaults()
		}
		defaults.Parse(os.Args)
		defaults.Usage()
	}

	if *versionFlag {
		fmt.Printf("dbbench %v, commit %v, built at %v\n", version, commit, date)
		os.Exit(0)
	}

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
	for i, b := range benchmarks {
		// check if we want to run this particular benchmark
		if !contains(toRun, "all") && !contains(toRun, b.Name) {
			continue
		}

		// Run the particular benchmark
		took := benchmark.Run(bencher, b, *iter, *threads)
		fmt.Printf("%v:\t%v\t%v\tns/op\n", b.Name, took, took.Nanoseconds()/int64(*iter))

		// Don't sleep after the last benchmark
		if i != len(benchmarks)-1 {
			time.Sleep(*sleep)
		}
	}
	fmt.Printf("total: %v\n", time.Since(startTotal))
}

func contains(options []string, want string) bool {
	for _, o := range options {
		if o == want {
			return true
		}
	}
	return false
}
