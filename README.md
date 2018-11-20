# sqlite

``` bash
go run main.go -db sqlite -iter 10000
``` 

# mysql

``` bash
docker run --name some-mysql -p 3306:3306 -d -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=dbbench mysql

go run main.go -db mysql -iter 5000 -port 3306 -user root -password root
``` 

# mariadb

``` bash
docker run --name some-mariadb -p 3306:3306 -d -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=dbbench mariadb 

go run main.go -db mariadb -iter 5000 -port 3306 -user root -password root
``` 

# postgres

``` bash
docker run -d -p 5432:5432 postgres

go run main.go -type pg -iter 1000 -port 36357 -user postgres -password example
``` 

# cockroach

``` bash
docker run -d -p 26257:26257 -p 8080:8080 cockroachdb/cockroach:latest start --iternsecure

go run main.go -type cr -iter 100 -port 26257 -user root
```

# cassandra

``` bash
docker run --name some-cassandra -p 9042:9042 -d cassandra:latest

go run main.go -db cassandra -iter 5000 -port 9042
```

# scylladb

``` bash
docker run --name some-scylla -p 9042:9042 -d scylladb/scylla

go run main.go -db scylla -iter 5000 -port 9042
``` 