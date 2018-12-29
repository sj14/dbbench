package benchmark

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseScript(t *testing.T) {
	type expect struct {
		benchmarks []Benchmark
		err        error
	}

	testCases := []struct {
		description string
		in          string
		expect      expect
	}{
		{
			description: "fail/no mode",
			in:          "\\benchmark",
			expect: expect{
				benchmarks: []Benchmark{},
				err:        ErrNoMode,
			},
		},
		{
			description: "fail/unknown mode",
			in:          "\\benchmark unknown-mode",
			expect: expect{
				benchmarks: []Benchmark{},
				err:        errors.New("failed to parse mode, neither 'once' nor 'loop': unknown-mode"),
			},
		},
		{
			description: "fail/missing name",
			in:          "\\benchmark once \\name",
			expect: expect{
				benchmarks: []Benchmark{},
				err:        ErrNoName,
			},
		},
		{
			description: "one statement",
			in:          "INSERT INTO ...;",
			expect: expect{
				benchmarks: []Benchmark{
					{Name: "(loop) line 1-1", Type: TypeLoop, Stmt: "INSERT INTO ...;"},
				},
			},
		},
		{
			description: "two statements",
			in: `
			INSERT INTO ...;
			DELETE FROM ...;
			`,
			expect: expect{
				benchmarks: []Benchmark{
					{Name: "(loop) line 1-4", Type: TypeLoop, Stmt: "INSERT INTO ...;\nDELETE FROM ...;"},
				},
			},
		},
		{
			description: "once/statement",
			in: `
			\benchmark once
			INSERT INTO ...;
			`,
			expect: expect{
				benchmarks: []Benchmark{
					{Name: "(once) line 3", Type: TypeOnce, Stmt: "INSERT INTO ...;"},
				},
			},
		},
		{
			description: "loop/once/statement",
			in: `
			\benchmark loop
			\benchmark once
			INSERT INTO ...;
			`,
			expect: expect{
				benchmarks: []Benchmark{
					{Name: "(once) line 4", Type: TypeOnce, Stmt: "INSERT INTO ...;"},
				},
			},
		},
		{
			description: "once/loop/statement",
			in: `
			\benchmark once
			\benchmark loop
			INSERT INTO ...;
			`,
			expect: expect{
				benchmarks: []Benchmark{
					{Name: "(loop) line 4-5", Type: TypeLoop, Stmt: "INSERT INTO ...;"},
				},
			},
		},
		{
			description: "once/once/statement",
			in: `
			\benchmark once
			\benchmark once
			INSERT INTO ...;
			`,
			expect: expect{
				benchmarks: []Benchmark{
					{Name: "(once) line 4", Type: TypeOnce, Stmt: "INSERT INTO ...;"},
				},
			},
		},
		{
			description: "loop/loop/statement",
			in: `
			\benchmark loop
			\benchmark loop
			INSERT INTO ...;
			`,
			expect: expect{
				benchmarks: []Benchmark{
					{Name: "(loop) line 4-5", Type: TypeLoop, Stmt: "INSERT INTO ...;"},
				},
			},
		},
		{
			description: "loop/two statements",
			in: `
			\benchmark loop
			INSERT INTO ...;
			DELETE FROM ...;
			`,
			expect: expect{
				benchmarks: []Benchmark{
					{Name: "(loop) line 3-5", Type: TypeLoop, Stmt: "INSERT INTO ...;\nDELETE FROM ...;"},
				},
			},
		},
		{
			description: "two loops/two statements",
			in: `
			\benchmark loop
			INSERT INTO ...;
			\benchmark loop
			UPDATE ...;
			`,
			expect: expect{
				benchmarks: []Benchmark{
					{Name: "(loop) line 3-3", Type: TypeLoop, Stmt: "INSERT INTO ...;"},
					{Name: "(loop) line 5-6", Type: TypeLoop, Stmt: "UPDATE ...;"},
				},
			},
		},
		{
			description: "once/two statements",
			in: `
			\benchmark once
			INSERT INTO ...;
			DELETE FROM ...;
			`,
			expect: expect{
				benchmarks: []Benchmark{
					{Name: "(once) line 3", Type: TypeOnce, Stmt: "INSERT INTO ...;"},
					{Name: "(once) line 4", Type: TypeOnce, Stmt: "DELETE FROM ...;"},
				},
			},
		},
		{
			description: "comment line",
			in: `
			-- MY COMMENT
			INSERT INTO ...;
			DELETE FROM ...;
			`,
			expect: expect{
				benchmarks: []Benchmark{
					{Name: "(loop) line 1-5", Type: TypeLoop, Stmt: "INSERT INTO ...;\nDELETE FROM ...;"},
				},
			},
		},
		{
			description: "inline comment",
			in: `
			INSERT INTO ...; -- MY COMMENT
			DELETE FROM ...;
			`,
			expect: expect{
				benchmarks: []Benchmark{
					{Name: "(loop) line 1-4", Type: TypeLoop, Stmt: "INSERT INTO ...; -- MY COMMENT\nDELETE FROM ...;"},
				},
			},
		},
		{
			description: "full example",
			in: `
			-- create table
			\benchmark once
			CREATE TABLE ...;
			
			-- how long takes an insert and delete?
			\benchmark loop
			INSERT INTO ...;
			DELETE FROM ...; 
			
			-- delete table
			\benchmark once
			DROP TABLE ...;
			`,
			expect: expect{
				benchmarks: []Benchmark{
					{Name: "(once) line 4", Type: TypeOnce, Stmt: "CREATE TABLE ...;"},
					{Name: "(loop) line 8-11", Type: TypeLoop, Stmt: "INSERT INTO ...;\nDELETE FROM ...;"},
					{Name: "(once) line 13", Type: TypeOnce, Stmt: "DROP TABLE ...;"},
				},
			},
		},
		{
			description: "set names",
			in: `
			\benchmark loop \name insert
			INSERT INTO ...;
			\benchmark loop \name update
			UPDATE ...;
			`,
			expect: expect{
				benchmarks: []Benchmark{
					{Name: "(loop) insert", Type: TypeLoop, Stmt: "INSERT INTO ...;"},
					{Name: "(loop) update", Type: TypeLoop, Stmt: "UPDATE ...;"},
				},
			},
		},
		{
			description: "loop/set 2/3 names",
			in: `
			\benchmark loop \name insert
			INSERT INTO ...;

			\benchmark loop
			UPDATE ...;
			
			\benchmark loop \name delete
			DELETE ...;
			`,
			expect: expect{
				benchmarks: []Benchmark{
					{Name: "(loop) insert", Type: TypeLoop, Stmt: "INSERT INTO ...;"},
					{Name: "(loop) line 6-7", Type: TypeLoop, Stmt: "UPDATE ...;"},
					{Name: "(loop) delete", Type: TypeLoop, Stmt: "DELETE ...;"},
				},
			},
		},
		{
			description: "once/set 2/3 names",
			in: `
			\benchmark once \name insert
			INSERT INTO ...;

			\benchmark once
			UPDATE ...;
			
			\benchmark once \name delete
			DELETE ...;
			`,
			expect: expect{
				benchmarks: []Benchmark{
					{Name: "(once) insert", Type: TypeOnce, Stmt: "INSERT INTO ...;"},
					{Name: "(once) line 6", Type: TypeOnce, Stmt: "UPDATE ...;"},
					{Name: "(once) delete", Type: TypeOnce, Stmt: "DELETE ...;"},
				},
			},
		},
		{
			description: "parallel",
			in: `
			\benchmark loop \parallel
			INSERT INTO ...;
			`,
			expect: expect{
				benchmarks: []Benchmark{
					{Name: "(loop) line 3-4", Type: TypeLoop, Parallel: true, Stmt: "INSERT INTO ...;"},
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.description, func(t *testing.T) {
			r := strings.NewReader(tt.in)

			// act
			got, err := ParseScript(r)
			require.Equal(t, tt.expect.err, err)

			// assert
			require.Equal(t, tt.expect.benchmarks, got)
		})
	}
}
