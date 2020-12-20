# dbbench

![Action](https://github.com/sj14/dbbench/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/sj14/dbbench)](https://goreportcard.com/report/github.com/sj14/dbbench)
[![Coverage Status](https://coveralls.io/repos/github/sj14/dbbench/badge.svg?branch=master)](https://coveralls.io/github/sj14/dbbench?branch=master)

## Table of Contents

- [Description](#Description)
- [Example](#example)
- [Installation](#installation)
- [Supported Databases](#Supported-Databases-/-Driver)
- [Usage](#usage)
- [Custom Scripts](#custom-scripts)
- [Troubeshooting](#troubleshooting)
- [Development](#development)
- [Acknowledgements](#Acknowledgements)

## Description

`dbbench` is a simple tool to benchmark or stress test databases. You can use the simple built-in benchmarks or run your own queries.  

**Attention**: This tool comes with no warranty. Don't run it on production databases.

## Example

``` text
$ dbbench postgres --user postgres --pass example --iter 100000
inserts 6.199670776s    61996   ns/op
updates 7.74049898s     77404   ns/op
selects 2.911541197s    29115   ns/op
deletes 5.999572479s    59995   ns/op
total: 22.85141994s
```

## Installation

### Precompiled Binaries

Binaries are available for all major platforms. See the [releases](https://github.com/sj14/dbbench/releases) page. Unfortunately, `cgo` is disabled for these builds, which means there is *no SQLite support* ([#1](https://github.com/sj14/dbbench/issues/1)).

### Homebrew

Using the [Homebrew](https://brew.sh/) package manager for macOS:

``` text
brew install sj14/tap/dbbench
```

### Manually

It's also possible to install the current development snapshot with `go get` (not recommended):

``` text
go get -u github.com/sj14/dbbench/cmd/dbbench
```

## Supported Databases / Driver

Databases | Driver
----------|-----------
Cassandra and compatible databases (e.g. ScyllaDB) | github.com/gocql/gocql
MS SQL and compatible databases (no built-in benchmarks yet) | github.com/denisenkom/go-mssqldb
MySQL and compatible databases (e.g. MariaDB and TiDB) | github.com/go-sql-driver/mysql
PostgreSQL and compatible databases (e.g. CockroachDB) | github.com/lib/pq
SQLite3 and compatible databases | github.com/mattn/go-sqlite3

## Usage

``` text
Available subcommands:
        cassandra|cockroach|mssql|mysql|postgres|sqlite
        Use 'subcommand --help' for all flags of the specified command.
Generic flags for all subcommands:
      --clean            only cleanup benchmark data, e.g. after a crash
      --iter int         how many iterations should be run (default 1000)
      --noclean          keep benchmark data
      --noinit           do not initialize database and tables, e.g. when only running own script
      --run string       only run the specified benchmarks, e.g. "inserts deletes" (default "all")
      --script string    custom sql file to execute
      --sleep duration   how long to pause after each single benchmark (valid units: ns, us, ms, s, m, h)
      --threads int      max. number of green threads (iter >= threads > 0) (default 25)
      --version          print version information
```

## Custom Scripts

You can run your own SQL statements with the `--script` flag. You can use the auto-generate tables. Beware the file size as it will be completely loaded into memory.

The script must contain valid SQL statements for your database.

There are some built-in variables and functions which can be used in the script. It's using the golang [template engine](https://golang.org/pkg/text/template/) which uses the delimiters `{{` and `}}`. Functions are executed with the `call` command and arguments are passed after the function name.

### Benchmark Settings

A new benchmark is created with the `\benchmark` keyword, followed by either `once` or `loop`. Optional parameters can be added afterwards in the same line.

The the usage description and the example subsection for more information.

Usage                     | Description                                   |
--------------------------|-----------------------------------------------|
`\benchmark once`                | Execute the following statements (lines) only once (e.g. to create and delete tables).
`\benchmark loop`                | Default mode. Execute the following statements (lines) in a loop. Executes them one after another and then starts a new iteration. Add another `\benchmark loop` to start another benchmark of statements.
`\name insert`              | Set a custom name for the DB statement(s), which will be output instead the line numbers (`insert` is an examplay name).

### Statement Substitutions

Usage                     | Description                                   |
--------------------------|-----------------------------------------------|
`{{.Iter}}`                 | The iteration counter. Will return `1` when `\benchmark once`.
`{{call .Seed 42}}`         | [godoc](https://golang.org/pkg/math/rand/#Seed) (`42` is an examplary seed)
`{{call .RandInt63}}`       | [godoc](https://golang.org/pkg/math/rand/#Int63)
`{{call .RandInt63n 9999}}` | [godoc](https://golang.org/pkg/math/rand/#Int63n) (`9999` is an examplary upper limit)
`{{call .RandFloat32}}`     | [godoc](https://golang.org/pkg/math/rand/#Float32)  
`{{call .RandFloat64}}`     | [godoc](https://golang.org/pkg/math/rand/#Float64)
`{{call .RandExpFloat64}}`  | [godoc](https://golang.org/pkg/math/rand/#ExpFloat64)
`{{call .RandNormFloat64}}` | [godoc](https://golang.org/pkg/math/rand/#NormFloat64)

### Example

Exemplary `sqlite_bench.sql` file:

``` sql
-- Create table
\benchmark once \name init
CREATE TABLE dbbench_simple (id INT PRIMARY KEY, balance DECIMAL);

-- How long takes an insert and delete?
\benchmark loop \name single
INSERT INTO dbbench_simple (id, balance) VALUES({{.Iter}}, {{call .RandInt63}});
DELETE FROM dbbench_simple WHERE id = {{.Iter}}; 

-- How long takes it in a single transaction?
\benchmark loop \name batch
BEGIN TRANSACTION;
INSERT INTO dbbench_simple (id, balance) VALUES({{.Iter}}, {{call .RandInt63}});
DELETE FROM dbbench_simple WHERE id = {{.Iter}}; 
COMMIT;

-- Delete table
\benchmark once \name clean
DROP TABLE dbbench_simple;
```

In this script, we create and delete the table manually, thus we will pass the `--noinit` and `--noclean` flag, which would otherwise create this default table for us:

``` text
dbbench sqlite --script scripts/sqlite_bench.sql --iter 5000 --noinit --noclean
```

output:

``` text
(once) init:    3.404784ms      3404784 ns/op
(loop) single:  10.568390874s   2113678 ns/op
(loop) batch:   5.739021596s    1147804 ns/op
(once) clean:   1.065703ms      1065703 ns/op
total: 16.312319959s
```

## Troubleshooting

**Error message**

``` text
failed to insert: UNIQUE constraint failed: dbbench_simple.id
```

**Description**
The previous data wasn't removed (e.g. because the benchmark was canceled). Try to run the same command again, but with the `--clean` flag attached, which will remove the old data. Then run the original command again.

---

**Error message**

``` text
failed to create table: Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub
```

**Description**  
Currently, the released binary builds don't contain SQLite support. You have to compile dbbench manually, either from the particular release source code (recommended) or from the current master branch (not recommended).

## Development

Below are some examples how to run different databases and the equivalent call of `dbbench` for testing/developing.

### Cassandra

``` text
docker run --name dbbench-cassandra -p 9042:9042 -d cassandra:latest
```

``` text
dbbench cassandra
```

### CockroachDB

``` text
# port 8080 is the webinterface (optional)
docker run --name dbbench-cockroach -d -p 26257:26257 -p 8080:8080 cockroachdb/cockroach:latest start --insecure
```

``` text
dbbench cockroach
```

### Microsoft SQL Server

``` text
docker run -e 'ACCEPT_EULA=Y' -e 'SA_PASSWORD=yourStrong(!)Password' -p 1433:1433 -d microsoft/mssql-server-linux
```

``` text
dbbench mssql -user sa -pass 'yourStrong(!)Password'
```

### MariaDB

``` text
docker run --name dbbench-mariadb -p 3306:3306 -d -e MYSQL_ROOT_PASSWORD=root mariadb
```

``` text
dbbench mariadb
```

### MySQL

``` text
docker run --name dbbench-mysql -p 3306:3306 -d -e MYSQL_ROOT_PASSWORD=root mysql
```

``` text
dbbench mysql
```

### PostgreSQL

``` text
docker run --name dbbench-postgres -p 5432:5432 -d postgres
```

``` text
dbbench postgres --user postgres --pass example
```

### ScyllaDB

``` text
docker run --name dbbench-scylla -p 9042:9042 -d scylladb/scylla
```

``` text
dbbench scylla
```

### SQLite

``` text
dbbench sqlite
```

### TiDB

``` text
git clone https://github.com/pingcap/tidb-docker-compose.git
```

``` text
cd tidb-docker-compose && docker-compose pull
```

``` text
docker-compose up -d
```

``` text
dbbench tidb --pass '' --port 4000
```

## Acknowledgements

Thanks to the authors of Go and those of the directly and indirectly used libraries, especially the driver developers. It wouldn't be possible without all your work.

This tool was highly inspired by the snippet from user [Fale](https://github.com/cockroachdb/cockroach/issues/23061#issue-300012178) and the tool [pgbench](https://www.postgresql.org/docs/current/pgbench.html). Later, also inspired by [MemSQL's dbbench](https://github.com/memsql/dbbench) which had the name and a similar idea before.