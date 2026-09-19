package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/sj14/dbbench/benchmark"
	"github.com/sj14/dbbench/databases"
	"github.com/spf13/pflag"
	_ "modernc.org/sqlite"
	_ "turso.tech/database/tursogo"
)

var (
	version = "dev version"
	commit  = "none"
	date    = "unknown"
)

func main() {
	os.Exit(run())
}

func run() int {
	var (
		// Default set of flags, available for all subcommands (benchmark options).
		defaultFlags = pflag.NewFlagSet("defaults", pflag.ExitOnError)
		iter         = defaultFlags.Int("iter", 1000, "how many iterations should be run")
		threads      = defaultFlags.Int("threads", 25, "max. number of green threads (iter >= threads > 0)")
		sleep        = defaultFlags.Duration("sleep", 0, "how long to pause after each single benchmark (valid units: ns, us, ms, s, m, h)")
		nosetup      = defaultFlags.Bool("noinit", false, "do not initialize database and tables, e.g. when only running own script")
		clean        = defaultFlags.Bool("clean", false, "only cleanup benchmark data, e.g. after a crash")
		noclean      = defaultFlags.Bool("noclean", false, "keep benchmark data")
		versionFlag  = defaultFlags.Bool("version", false, "print version information")
		runBench     = defaultFlags.String("run", "all", "only run the specified benchmarks, e.g. \"inserts deletes\"")
		scriptname   = defaultFlags.String("script", "", "custom sql file to execute")

		// Connection flags, applicable for most databases (not sqlite).
		connFlags = pflag.NewFlagSet("conn", pflag.ExitOnError)
		host      = connFlags.String("host", "localhost", "address of the server")
		port      = connFlags.Int("port", 0, "port of the server (0 -> db defaults)")
		user      = connFlags.String("user", "root", "user name to connect with the server")
		pass      = connFlags.String("pass", "root", "password to connect with the server")

		// Max. connections, applicable for most databases (not cassandra, sqlite).
		maxconnsFlags = pflag.NewFlagSet("conns", pflag.ExitOnError)
		maxconns      = maxconnsFlags.Int("conns", 0, "max. number of open connections")

		// GCP specific application flags (for Spanner)
		gcpFlags        = pflag.NewFlagSet("gcp", pflag.ExitOnError)
		instanceID      = gcpFlags.String("instance", "", "ID of the Spanner instance")
		projectID       = gcpFlags.String("project", "", "GCP project ID")
		databaseID      = gcpFlags.String("database", "", "ID of the Spanner Database")
		credentialsFile = gcpFlags.String("credentials", "", "optional file containing GCP credentials; defaults to application default credentials")

		// Flag sets for each database. DB specific flags are set in the switch statement below.
		cassandraFlags = pflag.NewFlagSet("cassandra", pflag.ExitOnError)
		cockroachFlags = pflag.NewFlagSet("cockroach", pflag.ExitOnError)
		mssqlFlags     = pflag.NewFlagSet("mssql", pflag.ExitOnError)
		mysqlFlags     = pflag.NewFlagSet("mysql", pflag.ExitOnError)
		postgresFlags  = pflag.NewFlagSet("postgres", pflag.ExitOnError)
		sqliteFlags    = pflag.NewFlagSet("sqlite", pflag.ExitOnError)
		spannerFlags   = pflag.NewFlagSet("spanner", pflag.ExitOnError)
		tursoFlags     = pflag.NewFlagSet("turso", pflag.ExitOnError)
	)

	defaultFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Available subcommands:\n\tcassandra|cockroach|mssql|mysql|postgres|sqlite|spanner|turso\n")
		fmt.Fprintf(os.Stderr, "\tUse 'subcommand --help' for all flags of the specified command.\n")
		fmt.Fprintf(os.Stderr, "Generic flags for all subcommands:\n")
		defaultFlags.PrintDefaults()
	}

	// No comamnd given. Print usage help and exit.
	if len(os.Args) < 2 {
		defaultFlags.Usage()
		return 1
	}

	var bencher benchmark.Bencher
	switch os.Args[1] {
	case "postgres":
		postgresFlags.AddFlagSet(defaultFlags)
		postgresFlags.AddFlagSet(connFlags)
		postgresFlags.AddFlagSet(maxconnsFlags)
		if err := postgresFlags.Parse(os.Args[2:]); err != nil {
			log.Fatalf("failed to parse postgres flags: %v", err)
		}
		bencher = databases.NewPostgres(*host, *port, *user, *pass, *maxconns)
	case "cockroach":
		cockroachFlags.AddFlagSet(defaultFlags)
		cockroachFlags.AddFlagSet(connFlags)
		cockroachFlags.AddFlagSet(maxconnsFlags)
		if err := cockroachFlags.Parse(os.Args[2:]); err != nil {
			log.Fatalf("failed to parse cockroach flags: %v", err)
		}
		bencher = databases.NewCockroach(*host, *port, *user, *pass, *maxconns)
	case "cassandra", "scylla":
		cassandraFlags.AddFlagSet(defaultFlags)
		cassandraFlags.AddFlagSet(connFlags)
		if err := cassandraFlags.Parse(os.Args[2:]); err != nil {
			log.Fatalf("failed to parse cassandra flags: %v", err)
		}
		bencher = databases.NewCassandra(*host, *port, *user, *pass)
	case "mysql", "mariadb", "tidb":
		mysqlFlags.AddFlagSet(defaultFlags)
		mysqlFlags.AddFlagSet(connFlags)
		mysqlFlags.AddFlagSet(maxconnsFlags)
		if err := mysqlFlags.Parse(os.Args[2:]); err != nil {
			log.Fatalf("failed to parse mysql flags: %v", err)
		}
		bencher = databases.NewMySQL(*host, *port, *user, *pass, *maxconns)
	case "mssql":
		mssqlFlags.AddFlagSet(defaultFlags)
		mssqlFlags.AddFlagSet(connFlags)
		mssqlFlags.AddFlagSet(maxconnsFlags)
		if err := mssqlFlags.Parse(os.Args[2:]); err != nil {
			log.Fatalf("failed to parse mssql flags: %v", err)
		}
		bencher = databases.NewMSSQL(*host, *port, *user, *pass, *maxconns)
	case "sqlite":
		sqliteFlags.AddFlagSet(defaultFlags)
		path := sqliteFlags.String("path", "dbbench.sqlite", "database file (sqlite only)")
		if err := sqliteFlags.Parse(os.Args[2:]); err != nil {
			log.Fatalf("failed to parse sqlite flags: %v", err)
		}
		bencher = databases.NewSQLite(*path)
	case "spanner":
		spannerFlags.AddFlagSet(defaultFlags)
		spannerFlags.AddFlagSet(gcpFlags)

		if err := spannerFlags.Parse(os.Args[2:]); err != nil {
			log.Fatalf("failed to parse spanner flags: %v", err)
		}
		bencher = databases.NewSpanner(*projectID, *instanceID, *databaseID, *credentialsFile)
	case "turso":
		tursoFlags.AddFlagSet(defaultFlags)
		tursoFlags.AddFlagSet(maxconnsFlags)
		path := tursoFlags.String("path", "dbbench.turso", "database file (Turso only)")
		if err := tursoFlags.Parse(os.Args[2:]); err != nil {
			log.Fatalf("failed to parse Turso flags: %v", err)
		}
		bencher = databases.NewTurso(*path, *maxconns)
	default:
		if err := defaultFlags.Parse(os.Args[1:]); err != nil {
			log.Fatalf("failed to parse default flags: %v", err)
		}

		// Only show version information and exit.
		if *versionFlag {
			fmt.Printf("dbbench %v, commit %v, built at %v\n", version, commit, date)
			return 0
		}

		// Command not recognized. Print usage help and exit.
		defaultFlags.Usage()
		return 1
	}

	// only clean old data when clean flag is set
	if *clean {
		bencher.Cleanup()
		fmt.Println("cleaned data")
		return 0
	}

	// setup database
	if !*nosetup {
		bencher.Setup()
	}

	// only cleanup benchmark data when noclean flag is not set
	if !*noclean && !*nosetup {
		defer bencher.Cleanup()
	}

	// we need at least one thread
	if *threads == 0 {
		*threads = 1
		fmt.Println("increased to 1 thread")
	}

	// can't have more threads than iterations
	if *threads > *iter {
		*threads = *iter
	}

	var benchmarks []benchmark.Benchmark
	if *scriptname != "" {
		dat, err := os.ReadFile(*scriptname)
		if err != nil {
			log.Fatalf("failed to read file: %v", err)
		}
		buf := bytes.NewBuffer(dat)
		benchmarks, err = benchmark.ParseScript(buf)
		if err != nil {
			log.Fatalf("failed to parse script: %v\n", err)
		}
	} else {
		benchmarks = bencher.Benchmarks()
		if len(benchmarks) == 0 {
			log.Printf("no built-in benchmarks available; use --script")
			return 1
		}
	}

	// split benchmark names when "-run 'bench0 bench1 ...'" flag was used
	toRun := strings.Split(*runBench, " ")

	startTotal := time.Now()

	// notify channel for SIGINT (ctrl-c)
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)

	for i, b := range benchmarks {
		select {
		case <-sigchan:
			// got SIGINT, stop benchmarking
			printTotal(startTotal)
			return 130
		default:
			// check if we want to run this particular benchmark
			if !contains(toRun, "all") && !contains(toRun, b.Name) {
				continue
			}

			// run the particular benchmark
			results, err := benchmark.Run(bencher, b, *iter, *threads)
			if err != nil {
				log.Printf("benchmark %q failed: %v", b.Name, err)
				return 1
			}

			took := results.Duration
			// execution in ns for mode once
			nsPerOp := took.Nanoseconds()

			// execution in ns/op for mode loop
			if b.Type == benchmark.TypeLoop {
				nsPerOp /= int64(*iter)
			}

			fmt.Printf(`%v (%vx) took: %v 
avg: %v, min: %v, max: %v
%v ops/s
%v ns/op

`,
				b.Name,
				results.TotalExecutionCount,
				took,
				results.Avg(),
				results.Min,
				results.Max,
				float64(results.TotalExecutionCount)/results.Duration.Seconds(),
				nsPerOp)

			// Don't sleep after the last benchmark
			if i != len(benchmarks)-1 {
				time.Sleep(*sleep)
			}
		}
	}
	printTotal(startTotal)
	return 0
}

func printTotal(startTotal time.Time) {
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
