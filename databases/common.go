package databases

import (
	"database/sql"
	"log"
)

// Benchmark contains a benchmark func and its name.
type Benchmark struct {
	Name string
	Func func(int)
}

func mustExec(result sql.Result, err error, name string) {
	if err != nil {
		log.Fatalf("failed to %v: %v\n", name, err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		log.Fatalf("failed to get %v rows: %v\n", name, err)
	}
	if rows == 0 {
		log.Fatalf("no rows %ved\n", name)
	}
}
