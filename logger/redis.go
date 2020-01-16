// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package logger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

const (
	cstRedisQueueName = "condensat.bank.logger.queue"
)

var (
	// ErrRedisFailed
	ErrRedisFailed = errors.New("Redis Failed")
)

type RedisLogger struct {
	rdb *redis.Client
}

type RedisOptions struct {
	HostName string
	Port     int
}

func NewRedisLogger(options RedisOptions) *RedisLogger {
	return &RedisLogger{
		rdb: redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("%s:%d", options.HostName, options.Port),
		}),
	}
}

// Write implements io.Writer interface
func (r *RedisLogger) Write(entry []byte) (int, error) {
	_, err := r.rdb.RPush(cstRedisQueueName, entry).Result()
	if err != nil {
		// print missed entry and exit
		print(string(entry))
		panic(err)
	}
	return len(entry), nil
}

// Grab entries from redis
func (r *RedisLogger) Grab(ctx context.Context) {
	log.SetLevel(contextLevel(ctx))
	log.SetOutput(os.Stderr)

	entryChan := make(chan []byte)

	go r.pullRedisEntries(ctx, entryChan, 1024, 20*time.Millisecond)
	go r.processEntries(ctx, entryChan)

	<-ctx.Done()
}

// pullRedisEntries publish entries from redis to entryChan
func (r *RedisLogger) pullRedisEntries(ctx context.Context, entryChan chan<- []byte, bulkSize int64, sleep time.Duration) {
	rdb := r.rdb
	for {
		// check for entries
		count, err := rdb.LLen(cstRedisQueueName).Result()
		if err != nil {
			panic(err)
		}
		if count == 0 {
			log.Tracef("Not log entry")
			time.Sleep(sleep)
			continue
		}

		// fetch bulk entries
		entries, err := rdb.LRange(cstRedisQueueName, 0, bulkSize-1).Result()
		if err != nil {
			panic(err)
		}

		// process entries
		for _, entry := range entries {
			entryChan <- []byte(entry)

			// entry processed, remove it from redis
			_, err = rdb.LPop(cstRedisQueueName).Result()
			if err != nil {
				panic(err)
			}
		}
		log.
			WithField("Count", len(entries)).
			Debug("Got entries")
	}
}

// processEntries consume log entries from entryChan
func (r *RedisLogger) processEntries(ctx context.Context, entryChan <-chan []byte) {
	for {
		select {
		case data := <-entryChan:
			var entry interface{}
			err := json.Unmarshal([]byte(data), &entry)
			if err != nil {
				continue
			}
			// print to stdout for now
			fmt.Fprint(os.Stdout, string(data))
		case <-ctx.Done():
		}
	}
}
