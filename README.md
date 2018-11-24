# Sample Usage

## Flags

``` bash
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
  -pass string
        password to connect with the server (default "root")
  -port int
        port of the server
  -run string
        only run the specified benchmarks, e.g. "inserts deletes" (default "all")
  -threads int
        max. number of green threads (default 25)
  -type string
        database to use (sqlite|mariadb|mysql|postgres|cockroach|cassandra|scylla)
  -user string
        user name to connect with the server (default "root")
``` 

## Scripts

`sqlite_bench.sql`: 
``` sql 
BEGIN TRANSACTION;
INSERT INTO accounts VALUES(1, 1);
DELETE FROM accounts WHERE id = 1; 
COMMIT;
``` 

Dont' use any comments in the file, it will be transformed to a single line before execution:

``` sql 
BEGIN TRANSACTION; INSERT INTO accounts VALUES(1, 1); DELETE FROM accounts WHERE id = 1; COMMIT;
``` 

`go run main.go -type sqlite -script sqlite_bench.sql`

## 

Below are some examples how to run different databases with docker and the equivalent call of dbbench for testing/developing.

## SQLite

driver: sqlite3

``` bash
go run main.go -type sqlite
``` 

## MySQL

driver: mysql

``` bash
docker run --name some-mysql -p 3306:3306 -d -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=dbbench mysql

go run main.go -type mysql -port 3306 -user root -password root
``` 

## MariaDB

driver: mysql

``` bash
docker run --name some-mariadb -p 3306:3306 -d -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=dbbench mariadb 

go run main.go -type mariadb -port 3306 -user root -password root
``` 

## PostgreSQL

driver: pg

``` bash
docker run -d -p 5432:5432 postgres

go run main.go -type postgres -port 5432 -user postgres -password example
``` 

## CockroachDB

driver: pg

``` bash
# port 8080 is the webinterface (optional)
docker run --name some-cockroach -d -p 26257:26257 -p 8080:8080 cockroachdb/cockroach:latest start --iternsecure

go run main.go -type cockroach -port 26257 -user root
```

## Cassandra

driver: gocql

``` bash
docker run --name some-cassandra -p 9042:9042 -d cassandra:latest

go run main.go -type cassandra -port 9042
```

## ScyllaDB

driver: gocql

``` bash
docker run --name some-scylla -p 9042:9042 -d scylladb/scylla

go run main.go -type scylla -port 9042
``` 

# Troubleshooting

I get the following error:

```
failed to insert: UNIQUE constraint failed: accounts.id
exit status 1
``` 
The previous data wasn't removed (e.g. because the benchmark was canceled). Try to run the same command again, but with the `-clean` flag attached, which will remove the old data. Then run the original command again.