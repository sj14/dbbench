package databases

import (
	"database/sql"
	"log"
)

// BenchType determines if the particular benchmark should be run several times or only once.
type BenchType int

const (
	// Loop executes the benchmark several times.
	Loop BenchType = iota
	// Once executes the benchmark once.
	Once BenchType = iota
)

// Benchmark contains the benchmark name, its db statement and its type.
type Benchmark struct {
	Name string
	Type BenchType
	Stmt string
}

func mustExec(result sql.Result, err error, name string) {
	if err != nil {
		log.Fatalf("%v: failed: %v\n", name, err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		log.Fatalf("%v: failed to get rows: %v\n", name, err)
	}
	if rows == 0 {
		log.Fatalf("%v: no rows\n", name)
	}
}
