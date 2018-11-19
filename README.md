# mysql

``` bash
docker run --name some-mysql -p 3306:3306 -d -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=dbbench mysql

go run main.go -db mysql -i 5000 -host localhost -port 3306 -user root -password root
``` 

# mariadb

``` bash
docker run --name some-mariadb -p 3306:3306 -d -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=dbbench mariadb 

go run main.go -db mariadb -i 5000 -host localhost -port 3306 -user root -password root
``` 

# postgres

``` bash
docker run -d -p 5432:5432 postgres

go run main.go -type pg -i 1000 -host localhost -port 36357 -user postgres -password example
``` 

# cockroach

``` bash
docker run -d -p 26257:26257 -p 8080:8080 cockroachdb/cockroach:latest start --insecure

go run main.go -type cr -i 100 -host localhost -port 26257 -user root
```

# cassandra

``` bash
docker run --name some-cassandra -p 9042:9042 -d cassandra:latest

go run main.go -db cassandra -i 5000 -host localhost -port 9042
```

# scylladb

``` bash
docker run --name some-scylla -p 9042:9042 -d scylladb/scylla

go run main.go -db scylla -i 5000 -host localhost -port 9042
``` 