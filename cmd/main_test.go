package main

import (
	"testing"

	"github.com/sj14/dbbench/databases"
)

func TestExecBenchmark(t *testing.T) {
	sqlite := databases.NewSQLite("dbbench_test.sqlite")
	sqlite.Setup()
	defer sqlite.Cleanup()

	for _, b := range sqlite.Benchmarks() {
		execBenchmark(b, 100, 25)
	}
}

func TestExecScript(t *testing.T) {
	sqlite := databases.NewSQLite("dbbench_test.sqlite")
	sqlite.Setup()
	defer sqlite.Cleanup()

	execScript(sqlite, "sqlite_bench.sql", 100, 25)
}
