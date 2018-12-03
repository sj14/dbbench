# dbbench

## Table of Contents

1. [Description](#Description)
1. [Example](#example)
4. [Installation](#installation)
2. [Supported Databases](#Supported-Databases-/-Driver)
4. [Usage](#usage)
4. [Custom Scripts](#custom-scripts)
4. [Known Issues](#known-issues)
3. [TODO](#TODO)
3. [Troubeshooting](#troubleshooting)
4. [Development](#development)
4. [Acknowledgements](#Acknowledgements)

## Description

`dbbench` is a simple tool to benchmark or stress test a databases. You can use the simple built-in benchmarks or run your own queries.  

**Attention**: This tool comes with no warranty. Don't run it on production databases.


## Example

``` text
$ dbbench -type postgres -user postgres -pass example -iter 100000
inserts 6.199670776s    61996   ns/op
updates 7.74049898s     77404   ns/op
selects 2.911541197s    29115   ns/op
deletes 5.999572479s    59995   ns/op
total: 22.85141994s
```

## Installation

### Precompiled Binaries

Binaries are available for all major platforms. See the [releases](https://github.com/sj14/dbbench/releases) page. Unfortunately, CGO is disabled for these builds, which means they are not compiled with SQLite support.

### Homebrew

Using the [Homebrew](https://brew.sh/) package manager for macOS:

```
brew install sj14/tap/dbbench
``` 

### Manually

It's also possible to install the current development snapshot with `go get` (not recommended):

``` text
go get -u github.com/sj14/dbbench/cmd
go install github.com/sj14/dbbench/cmd
```

## Supported Databases / Driver

Databases | Driver
----------|-----------
SQLite3 and compatible databases | github.com/mattn/go-sqlite3
MySQL and compatible databases (e.g. MariaDB) | github.com/go-sql-driver/mysql
PostgreSQL and compatible databases (e.g. CockroachDB) | github.com/lib/pq
Cassandra and compatible databases (e.g. ScyllaDB) | github.com/gocql/gocql

## Usage

``` text
  -clean
        only cleanup benchmark data, e.g. after a crash
  -conns int
        max. number of open connections
  -host string
        address of the server (default "localhost")
  -iter int
        how many iterations should be run (default 1000)
  -noclean
        keep benchmark data
  -noinit
        do not initialize database and tables, e.g. when only running own script
  -pass string
        password to connect with the server (default "root")
  -path string
        database file (sqlite only) (default "dbbench.sqlite")
  -port int
        port of the server (0 -> db defaults)
  -run string
        only run the specified benchmarks, e.g. "inserts deletes" (default "all")
  -script string
        custom sql file to execute
  -sleep duration
        how long to pause after each single benchmark (valid units: ns, us, ms, s, m, h)
  -threads int
        max. number of green threads (iter >= threads > 0 (default 25)
  -type string
        database to use (sqlite|mysql|postgres|cockroach|cassandra)
  -user string
        user name to connect with the server (default "root")
  -version
        print version information
```

## Custom Scripts

You can run your own SQL statements with the `-script` flag. You can use the auto-generate tables. Beware the file size as it will be completely loaded into memory.


The script must contain valid SQL statements for your database.

There are some built-in variables and functions which can be used in the script. It's using the golang [template engine](https://golang.org/pkg/text/template/) which uses the delimiters `{{` and `}}`. Functions are executed with the `call` command and arguments are passed after the function name.

Usage                     | Description                                   |
--------------------------|-----------------------------------------------|
`\mode once`                | Execute the following statements only once (e.g. to create and delete tables).
`\mode loop`                | Execute the following statements in a loop. Executes them one after another and then starts a new iteration. Add another `\mode loop` to start another benchmark of statements.
`{{.Iter}}`                 | The iteration counter.
`{{call .Seed 42}}`         | [godoc](https://golang.org/pkg/math/rand/#Seed) (42 is an examplary seed)
`{{call .RandInt63}}`       | [godoc](https://golang.org/pkg/math/rand/#Int63) 
`{{call .RandInt63n 9999}}` | [godoc](https://golang.org/pkg/math/rand/#Int63n) (9999 is an examplary upper limit)
`{{call .RandFloat32}}`     | [godoc](https://golang.org/pkg/math/rand/#Float32)  
`{{call .RandFloat64}}`     | [godoc](https://golang.org/pkg/math/rand/#Float64)
`{{call .RandExpFloat64}}`  | [godoc](https://golang.org/pkg/math/rand/#ExpFloat64)
`{{call .RandNormFloat64}}` | [godoc](https://golang.org/pkg/math/rand/#NormFloat64)

Exemplary `sqlite_bench.sql` file:

``` sql
-- create table
\mode once
CREATE TABLE dbbench_simple (id INT PRIMARY KEY, balance DECIMAL);

-- how long takes an insert and delete?
\mode loop
INSERT INTO dbbench_simple (id, balance) VALUES({{.Iter}}, {{call .RandInt63}});
DELETE FROM dbbench_simple WHERE id = {{.Iter}}; 

-- low long takes it in a single transaction?
\mode loop
BEGIN TRANSACTION;
INSERT INTO dbbench_simple (id, balance) VALUES({{.Iter}}, {{call .RandInt63}});
DELETE FROM dbbench_simple WHERE id = {{.Iter}}; 
COMMIT;

-- delete table
\mode once
DROP TABLE dbbench_simple;
```

In this script, we create and delete the table manually, thus we will pass the `-noinit` and `-noclean` flag, which would otherwise create this default table for us:

``` text
dbbench -type sqlite -script scripts/sqlite_bench.sql -iter 1000 -noinit -noclean
```

output:
``` text
once: line 3:           3.781231ms      3781    ns/op
loop: line 7-10:        7.334864879s    7334864 ns/op
loop: line 12-17:       3.481569798s    3481569 ns/op
once: line 19:          3.903027ms      3903    ns/op
total: 10.824351356s
```

## Known Issues

- Releases are built without CGO support (no support for sqlite) [#1](https://github.com/sj14/dbbench/issues/1)

## TODO

- [ ] More benchmarks and database specific benchmarks
  - [x] ~~Relational DB specific (e.g. MySQL)~~
  - [ ] Non-relational DB specific (e.g. Cassandra)
- [ ] More databases
  - [ ] MSSQL
  - [ ] MongoDB
  - [ ] ...

## Troubleshooting

**Error message**

``` text
failed to insert: UNIQUE constraint failed: dbbench_simple.id
```

**Description**   
The previous data wasn't removed (e.g. because the benchmark was canceled). Try to run the same command again, but with the `-clean` flag attached, which will remove the old data. Then run the original command again.

---

**Error message**

``` text
failed to create table: Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub
```

**Description**  
Currently, the released binary builds don't contain SQLite support. You have to compile dbbench manually, either from the particular release source code (recommended) or from the current master branch (not recommended).

## Development

Below are some examples how to run different databases and the equivalent call of `dbbench` for testing/developing.

### SQLite

``` text
dbbench -type sqlite
```

### MySQL

``` text
docker run --name dbbench-mysql -p 3306:3306 -d -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=dbbench mysql
dbbench -type mysql
```

### MariaDB

``` text
docker run --name dbbench-mariadb -p 3306:3306 -d -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=dbbench mariadb
dbbench -type mariadb
```

### PostgreSQL

``` text
docker run --name dbbench-postgres -p 5432:5432 -d postgres
dbbench -type postgres -user postgres -pass example
```

### CockroachDB

``` text
# port 8080 is the webinterface (optional)
docker run --name dbbench-cockroach -d -p 26257:26257 -p 8080:8080 cockroachdb/cockroach:latest start --insecure
dbbench -type cockroach
```

### Cassandra

``` text
docker run --name dbbench-cassandra -p 9042:9042 -d cassandra:latest
dbbench -type cassandra
```

### ScyllaDB

``` text
docker run --name dbbench-scylla -p 9042:9042 -d scylladb/scylla
dbbench -type scylla
```

## Acknowledgements

Thanks to the authors of Go and those of the directly and indirectly used libraries, especially the driver developers. It wouldn't be possible without all your work.

This tool was highly inspired by the snippet from user [Fale](https://github.com/cockroachdb/cockroach/issues/23061#issue-300012178) and the tool [pgbench](https://www.postgresql.org/docs/current/pgbench.html).