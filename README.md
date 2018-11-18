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

