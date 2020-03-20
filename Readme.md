# Condensat Bank backend

This repository hold all the backend components for Condensat Bank.

## Logging system

### Start mariadb

Log can be stored into mariadb.

```bash
docker run --name mariadb-test -e MYSQL_RANDOM_ROOT_PASSWORD=yes -e MYSQL_USER=condensat -e MYSQL_PASSWORD=condensat -e MYSQL_DATABASE=condensat -v $(pwd)/tests/database/permissions.sql:/docker-entrypoint-initdb.d/permissions.sql:ro -p 3306:3306 -d mariadb:10.3
```

### Start redis

Redis is used as a cache for logging to avoid message loses.

``` bash
docker run --name redis-test -p 6379:6379 -d redis:5-alpine
```

### Start the log grabber
The log grabber fetch log entries from redis and display them.
Log entries are remove from redis after store


```bash
go run logger/cmd/grabber/main.go --log=debug > ../debug.log
```

### Start the log grabber with database
The log grabber fetch log entries from redis and store them to database.
Log entries are remove from redis after store


```bash
go run logger/cmd/grabber/main.go --log=debug --withDatabase=true
```

### Use RedisLogger

A logging component setup a RedisLogger and log normally.

```bash
go run logger/cmd/example/main.go --appName=Foo --log=debug
```

## Messaging system

Nats is used for internal messaging system between components.

### Start nats

``` bash
docker run --name nats-test -p 4222:4222 -d nats:2.1-alpine
```

## Unit testing

```bash
go test -v ./...
```