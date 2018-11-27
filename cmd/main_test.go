package main

import (
	"testing"

	"github.com/sj14/dbbench/databases"
)

func TestExecBenchmark(t *testing.T) {
	sqlite := databases.NewSQLite("dbbench_test.sqlite")
	sqlite.Setup()
	defer sqlite.Cleanup()
	benchmark(sqlite, "", "all", 100, 25)
}

func TestExecScript(t *testing.T) {
	sqlite := databases.NewSQLite("dbbench_test.sqlite")
	sqlite.Setup()
	defer sqlite.Cleanup()
	benchmark(sqlite, "../scripts/sqlite_bench.sql", "all", 100, 25)
}
