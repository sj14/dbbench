package benchmark

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseScript(t *testing.T) {
	// arrange
	testCases := []struct {
		description string
		in          string
		expect      []Benchmark
	}{
		{
			description: "one statement",
			in:          "INSERT INTO ...;",
			expect: []Benchmark{
				{Name: "loop: line 1-1", Type: TypeLoop, Stmt: "INSERT INTO ...;"},
			},
		},
		{
			description: "two statements",
			in: `
			INSERT INTO ...;
			DELETE FROM ...;
			`,
			expect: []Benchmark{
				{Name: "loop: line 1-4", Type: TypeLoop, Stmt: "INSERT INTO ...;\nDELETE FROM ...;"},
			},
		},
		{
			description: "once/statement",
			in: `
			\mode once
			INSERT INTO ...;
			`,
			expect: []Benchmark{
				{Name: "once: line 3", Type: TypeOnce, Stmt: "INSERT INTO ...;"},
			},
		},
		{
			description: "loop/once/statement",
			in: `
			\mode loop
			\mode once
			INSERT INTO ...;
			`,
			expect: []Benchmark{
				{Name: "once: line 4", Type: TypeOnce, Stmt: "INSERT INTO ...;"},
			},
		},
		{
			description: "once/loop/statement",
			in: `
			\mode once
			\mode loop
			INSERT INTO ...;
			`,
			expect: []Benchmark{
				{Name: "loop: line 4-5", Type: TypeLoop, Stmt: "INSERT INTO ...;"},
			},
		},
		{
			description: "once/once/statement",
			in: `
			\mode once
			\mode once
			INSERT INTO ...;
			`,
			expect: []Benchmark{
				{Name: "once: line 4", Type: TypeOnce, Stmt: "INSERT INTO ...;"},
			},
		},
		{
			description: "loop/loop/statement",
			in: `
			\mode loop
			\mode loop
			INSERT INTO ...;
			`,
			expect: []Benchmark{
				{Name: "loop: line 4-5", Type: TypeLoop, Stmt: "INSERT INTO ...;"},
			},
		},
		{
			description: "loop/two statements",
			in: `
			\mode loop
			INSERT INTO ...;
			DELETE FROM ...;
			`,
			expect: []Benchmark{
				{Name: "loop: line 3-5", Type: TypeLoop, Stmt: "INSERT INTO ...;\nDELETE FROM ...;"},
			},
		},
		{
			description: "two loops/two statements",
			in: `
			\mode loop
			INSERT INTO ...;
			\mode loop
			UPDATE ...;
			`,
			expect: []Benchmark{
				{Name: "loop: line 3-3", Type: TypeLoop, Stmt: "INSERT INTO ...;\n"},
				{Name: "loop: line 5-6", Type: TypeLoop, Stmt: "UPDATE ...;"},
			},
		},
		{
			description: "once/two statements",
			in: `
			\mode once
			INSERT INTO ...;
			DELETE FROM ...;
			`,
			expect: []Benchmark{
				{Name: "once: line 3", Type: TypeOnce, Stmt: "INSERT INTO ...;"},
				{Name: "once: line 4", Type: TypeOnce, Stmt: "DELETE FROM ...;"},
			},
		},
		{
			description: "comment line",
			in: `
			-- MY COMMENT
			INSERT INTO ...;
			DELETE FROM ...;
			`,
			expect: []Benchmark{
				{Name: "loop: line 1-5", Type: TypeLoop, Stmt: "INSERT INTO ...;\nDELETE FROM ...;"},
			},
		},
		{
			description: "inline comment",
			in: `
			INSERT INTO ...; -- MY COMMENT
			DELETE FROM ...;
			`,
			expect: []Benchmark{
				{Name: "loop: line 1-4", Type: TypeLoop, Stmt: "INSERT INTO ...; -- MY COMMENT\nDELETE FROM ...;"},
			},
		},
		{
			description: "full example",
			in: `
			-- create table
			\mode once
			CREATE TABLE ...;
			
			-- how long takes an insert and delete?
			\mode loop
			INSERT INTO ...;
			DELETE FROM ...; 
			
			-- delete table
			\mode once
			DROP TABLE ...;
			`,
			expect: []Benchmark{
				{Name: "once: line 4", Type: TypeOnce, Stmt: "CREATE TABLE ...;"},
				{Name: "loop: line 8-11", Type: TypeLoop, Stmt: "INSERT INTO ...;\nDELETE FROM ...;\n"},
				{Name: "once: line 13", Type: TypeOnce, Stmt: "DROP TABLE ...;"},
			},
		},
		{
			description: "name statement",
			in: `
			\name mybench
			INSERT INTO ...;`,
			expect: []Benchmark{
				{Name: "mybench", Type: TypeLoop, Stmt: "INSERT INTO ...;"},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.description, func(t *testing.T) {
			r := strings.NewReader(tt.in)

			// act
			got := ParseScript(r)

			// assert
			require.Equal(t, tt.expect, got)
		})
	}
}
