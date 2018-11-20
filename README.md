# Sample Usage

Here are some examples how to run different databases with docker and the equivalent call of dbbench for testing/developing.

## SQLite

``` bash
go run main.go -db sqlite
``` 

## MySQL

``` bash
docker run --name some-mysql -p 3306:3306 -d -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=dbbench mysql

go run main.go -db mysql -port 3306 -user root -password root
``` 

## MariaDB

``` bash
docker run --name some-mariadb -p 3306:3306 -d -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=dbbench mariadb 

go run main.go -db mariadb -port 3306 -user root -password root
``` 

## PostgreSQL

``` bash
docker run -d -p 5432:5432 postgres

go run main.go -type postgres -port 5432 -user postgres -password example
``` 

## CockroachDB

``` bash
# port 8080 is the webinterface (optional)
docker run --name some-cockroach -d -p 26257:26257 -p 8080:8080 cockroachdb/cockroach:latest start --iternsecure

go run main.go -type cockroach -port 26257 -user root
```

## Cassandra

``` bash
docker run --name some-cassandra -p 9042:9042 -d cassandra:latest

go run main.go -db cassandra -port 9042
```

## ScyllaDB

``` bash
docker run --name some-scylla -p 9042:9042 -d scylladb/scylla

go run main.go -db scylla -port 9042
``` 

# Troubleshooting

I get the following error:

```
failed to insert: UNIQUE constraint failed: accounts.id
exit status 1
``` 
The previous data wasn't removed (e.g. because the benchmark was canceled). Try to run the same command again, but with the `-clean` flag attached, which will remove the old data. Then run the original command again.