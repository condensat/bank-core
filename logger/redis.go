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

	"github.com/condensat/bank-core/logger/model"

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

	entryChan := make(chan [][]byte)

	go r.pullRedisEntries(ctx, entryChan, 256, 20*time.Millisecond)

	<-ctx.Done()
}

// pullRedisEntries publish entries from redis to entryChan
func (r *RedisLogger) pullRedisEntries(ctx context.Context, entryChan chan<- [][]byte, bulkSize int64, sleep time.Duration) {
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

		log.
			WithField("Count", len(entries)).
			Debug("Got entries")

		// process entries
		var data [][]byte
		for _, entry := range entries {
			data = append(data, []byte(entry))
		}
		r.processEntries(ctx, data)

		for range entries {
			// entry processed, remove it from redis
			_, err = rdb.LPop(cstRedisQueueName).Result()
			if err != nil {
				panic(err)
			}
		}
	}
}

// processEntries consume log entries from entryChan
func (r *RedisLogger) processEntries(ctx context.Context, datas [][]byte) {
	databaseLogger := contextDatabase(ctx)
	if databaseLogger != nil {
		var logEntries []*model.LogEntry
		for _, data := range datas {

			var entry interface{}
			err := json.Unmarshal(data, &entry)
			if err != nil {
				// not json, print to stdout
				fmt.Fprint(os.Stdout, string(data))
				continue
			}

			m := entry.(map[string]interface{})
			timestamp := time.Now().UTC().Round(time.Second)
			if ts, ok := m["time"]; ok {
				t, err := time.Parse(time.RFC3339, ts.(string))
				if err == nil {
					timestamp = t
					delete(m, "time")
				}
			}
			app := m["app"].(string)
			delete(m, "app")
			level := m["level"].(string)
			delete(m, "level")
			msg := m["msg"].(string)
			delete(m, "msg")
			d, err := json.Marshal(m)
			if err == nil {
				data = d
			}

			logEntries = append(logEntries, databaseLogger.CreateLogEntry(timestamp, app, level, msg, string(data)))
		}

		err := databaseLogger.AddLogEntries(logEntries)
		if err != nil {
			log.
				WithError(err).
				Error("Fail to Add log entries")
			return
		}
	} else {
		// print to stdout
		for _, data := range datas {
			fmt.Fprint(os.Stdout, string(data))
		}
	}
}
