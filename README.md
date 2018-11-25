# dbbench

`dbbench` is a simple tool to benchmark or stress test a databases. You can use the simple built-in benchmarks or run your own queries.  

**Attention**: This tool comes with no warranty. Don't run it on a production database or know what you do.

## Example

``` text
$ dbbench -type postgres -user postgres -pass example -iter 100000
inserts 6.199670776s    61996   ns/op
updates 7.74049898s     77404   ns/op
selects 2.911541197s    29115   ns/op
deletes 5.999572479s    59995   ns/op
total: 22.85141994s
```

## Supported Databases / Driver

Databases | Driver
----------|-----------
SQLite3 and compatible databases | [github.com/mattn/go-sqlite3]
MySQL and compatible databases (e.g. MariaDB) | [github.com/go-sql-driver/mysql]
PostgreSQL and compatible databases (e.g. CockroachDB) | [github.com/lib/pq]
Cassandra and compatible databases (e.g. ScyllaDB) | [github.com/gocql/gocql]

## TODO

- [ ] More and database specific benchmarks
  - [ ] Relational DB specific (e.g. MySQL)
  - [ ] Non-relational DB specific (e.g. Cassandra)
- [ ] More databases
  - [ ] MSSQL
  - [ ] MongoDB
  - [ ] ...

## Installation

### Precompiled binaries

Binaries are available for all major platforms. See the [releases](https://github.com/sj14/dbbench/releases) page.


### Manually

It's also possible to install the current development snapshot with `go get` (not recommended):

``` text
go get -u github.com/sj14/dbbench
```

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
  -threads int
        max. number of green threads (default 25)
  -type string
        database to use (sqlite|mariadb|mysql|postgres|cockroach|cassandra|scylla)
  -user string
        user name to connect with the server (default "root")
  -version
        print version information
```

## Scripts

You can run your own SQL statements with the `-script` flag. You can use the auto-generate tables. Beware the file size as it will be completely loaded into memory!

``` text
$ dbbench -type sqlite -script scripts/sqlite_bench.sql -iter 1000
custom script: 3.851557272s     3851557 ns/op
total: 3.85158506s
```

The script must only contain valid SQL statements for your database.

`sqlite_bench.sql`:

``` sql
INSERT INTO accounts VALUES(1, 1);
DELETE FROM accounts WHERE id = 1;
```

**Don't use comments** in the file, it will be transformed to a single line before execution:

``` sql
-- my comment uncommented everything INSERT INTO accounts VALUES(1, 1); DELETE FROM accounts WHERE id = 1;
```

## Examples

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

## Troubleshooting

I get the following error:

``` text
failed to insert: UNIQUE constraint failed: accounts.id
exit status 1
```

The previous data wasn't removed (e.g. because the benchmark was canceled). Try to run the same command again, but with the `-clean` flag attached, which will remove the old data. Then run the original command again.

## Acknowledgements

Thanks to the authors of Go and those of the directly and indirectly used libraries, especially the driver developers. It wouldn't be possible without all your work.

This tool was highly inspired by the snipped from user [Fale](https://github.com/cockroachdb/cockroach/issues/23061#issue-300012178) and the tool [pgbench](https://www.postgresql.org/docs/current/pgbench.html).