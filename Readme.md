# Condensat Bank backend

This repository hold all the backend components for Condensat Bank.

## Logging system

### Start redis

Redis is used as a cache for logging to avoid message loses.

``` bash
docker run --name redis-test -p 6379:6379 -d redis:5-alpine
```

### Start the log grabber
The log grabber fetch log entries from redis and store them.
Log entries are remove from redis after store


```bash
go run logger/cmd/grabber/main.go --log=debug > ../debug.log
```

### Use RedisLogger

A logging component setup a RedisLogger and log normally.

```bash
go run logger/cmd/example/main.go --appName=Foo --log=debug
```
