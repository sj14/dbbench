package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"text/template"
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

	benchmarks := []databases.Benchmark{}

	// Benchmark specified script
	if *scriptname != "" {
		dat, err := ioutil.ReadFile(*scriptname)
		if err != nil {
			log.Fatalf("failed to read file: %v", err)
		}
		buf := bytes.NewBuffer(dat)
		benchmarks = parseScript(buf)
		for _, b := range benchmarks {
			fmt.Printf("%+v\n", b)
		}
		fmt.Println()
	} else {
		benchmarks = bencher.Benchmarks()
	}

	// split benchmark names when "-run 'bench0 bench1 ...'" flag was used
	toRun := strings.Split(*runBench, " ")

	startTotal := time.Now()
	// select built-in benchmarks
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
		if b.Type == databases.Once {
			exec(bencher, t, i)
		} else {
			benchmark(t, bencher, *iter, *threads)
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

func parseScript(r io.Reader) []databases.Benchmark {
	s := bufio.NewScanner(r)
	benchmarks := []databases.Benchmark{}

	mode := databases.Loop
	loopStmt := ""
	loopStart := 1
	lineN := 1
	for ; s.Scan(); lineN++ {
		line := s.Text()

		// skip comments and empty lines
		if strings.HasPrefix(line, "--") || line == "" {
			continue
		}

		if strings.HasPrefix(line, "\\mode") {
			if strings.Contains(line, "once") {
				// once
				if mode == databases.Loop {
					if loopStmt != "" {
						// was loop before, flush loop statements
						benchmarks = append(benchmarks, databases.Benchmark{Name: fmt.Sprintf("loop: line %v-%v", loopStart, lineN-1), Type: databases.Loop, Stmt: loopStmt})
						loopStmt = ""
					}
				}
				mode = databases.Once
			} else if strings.Contains(line, "loop") {
				// loop
				if loopStmt != "" {
					// also was loop before, flush loop statements and start a new loop statement
					benchmarks = append(benchmarks, databases.Benchmark{Name: fmt.Sprintf("loop: line %v-%v", loopStart, lineN-1), Type: databases.Loop, Stmt: loopStmt})
					loopStmt = ""
				}
				mode = databases.Loop
				loopStart = lineN + 1
			} else {
				log.Fatalf("failed to parse mode: %v", line)
			}
			// don't append \mode line
			continue
		}

		switch mode {
		case databases.Once:
			// Once, append benchmark immediately.
			benchmarks = append(benchmarks, databases.Benchmark{Name: fmt.Sprintf("once: line %v", lineN), Type: databases.Once, Stmt: line})
		case databases.Loop:
			// Loop, but not finished yet, append only line.
			loopStmt += line + "\n"
		}
	}

	// reached the end of the file, append remaining loop statements to benchmark
	if loopStmt != "" {
		benchmarks = append(benchmarks, databases.Benchmark{Name: fmt.Sprintf("loop: line %v-%v", loopStart, lineN-1), Type: databases.Loop, Stmt: loopStmt})
	}

	return benchmarks
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

func benchmark(t *template.Template, bencher Bencher, iterations, goroutines int) {
	wg := &sync.WaitGroup{}
	wg.Add(goroutines)
	defer wg.Wait()

	for routine := 0; routine < goroutines; routine++ {
		from := ((iterations / goroutines) * routine) + 1
		to := (iterations / goroutines) * (routine + 1)

		go func(gofrom, togo int) {
			defer wg.Done()

			for i := gofrom; i <= togo; i++ {
				exec(bencher, t, i)
			}
		}(from, to)
	}
}

// TODO: find better names/structure of functions
func exec(bencher Bencher, t *template.Template, i int) {
	sb := &strings.Builder{}

	data := struct {
		Iter            int
		Seed            func(int64)
		RandInt63       func() int64
		RandInt63n      func(int64) int64
		RandFloat32     func() float32
		RandFloat64     func() float64
		RandExpFloat64  func() float64
		RandNormFloat64 func() float64
	}{
		Iter:            i,
		Seed:            rand.Seed,
		RandInt63:       rand.Int63,
		RandInt63n:      rand.Int63n,
		RandFloat32:     rand.Float32,
		RandFloat64:     rand.Float64,
		RandExpFloat64:  rand.ExpFloat64,
		RandNormFloat64: rand.NormFloat64,
	}
	if err := t.Execute(sb, data); err != nil {
		log.Fatalf("failed to execute template: %v", err)
	}
	bencher.Exec(sb.String())
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

func contains(options []string, want string) bool {
	for _, o := range options {
		if o == want {
			return true
		}
	}
	return false
}
