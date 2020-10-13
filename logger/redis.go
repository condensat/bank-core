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

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger/model"

	log "github.com/sirupsen/logrus"
)

const (
	cstRedisQueueName = "condensat.bank.logger.queue"
)

var (
	ErrInvalidCache = errors.New("Invalid Cache")
)

type RedisLogger struct {
	cache bank.Cache
}

func NewRedisLogger(ctx context.Context) *RedisLogger {
	cache := appcontext.Cache(ctx)
	if cache == nil {
		panic(ErrInvalidCache)
	}
	return &RedisLogger{
		cache: cache,
	}
}

// Write implements io.Writer interface
func (r *RedisLogger) Write(entry []byte) (int, error) {
	rdb := cache.ToRedis(r.cache)
	_, err := rdb.RPush(context.Background(), cstRedisQueueName, entry).Result()
	if err != nil {
		// print missed entry and exit
		print(string(entry))
		panic(err)
	}
	return len(entry), nil
}

// Grab entries from redis
func (r *RedisLogger) Grab(ctx context.Context) {
	log.SetLevel(appcontext.Level(ctx))
	log.SetOutput(os.Stderr)

	entryChan := make(chan [][]byte)

	go r.pullRedisEntries(ctx, entryChan, 256, time.Second)

	<-ctx.Done()
}

// pullRedisEntries publish entries from redis to entryChan
func (r *RedisLogger) pullRedisEntries(ctx context.Context, entryChan chan<- [][]byte, bulkSize int64, sleep time.Duration) {
	rdb := cache.ToRedis(r.cache)
	for {
		// check for entries
		count, err := rdb.LLen(ctx, cstRedisQueueName).Result()
		if err != nil {
			panic(err)
		}
		if count == 0 {
			log.Tracef("Not log entry")
			<-time.After(sleep)
			continue
		}

		// fetch bulk entries
		entries, err := rdb.LRange(ctx, cstRedisQueueName, 0, bulkSize-1).Result()
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
			_, err = rdb.LPop(ctx, cstRedisQueueName).Result()
			if err != nil {
				panic(err)
			}
		}
	}
}

// processEntries consume log entries from entryChan
func (r *RedisLogger) processEntries(ctx context.Context, datas [][]byte) {
	ctxLogger := appcontext.Logger(ctx)
	if ctxLogger != nil {
		var logEntries []*model.LogEntry
		for _, data := range datas {

			var entry interface{}
			{
				err := json.Unmarshal(data, &entry)
				if err != nil {
					// not json, print to stdout
					fmt.Fprint(os.Stdout, string(data))
					continue
				}
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

			var userID uint64
			if uid, ok := m["UserID"].(float64); ok {
				delete(m, "UserID")
				userID = uint64(uid)
			}

			var sessionID string
			if sid, ok := m["SessionID"].(string); ok {
				delete(m, "SessionID")
				sessionID = sid
			}

			var method string
			if mtd, ok := m["Method"].(string); ok {
				delete(m, "Method")
				method = mtd
			}

			var err string
			if e, ok := m["error"].(string); ok {
				delete(m, "error")
				err = e
			}

			msg := m["msg"].(string)
			delete(m, "msg")

			if d, err := json.Marshal(m); err == nil {
				data = d
			}

			logEntries = append(logEntries, ctxLogger.CreateLogEntry(timestamp, app, level, userID, sessionID, method, err, msg, string(data)))
		}

		err := ctxLogger.AddLogEntries(logEntries)
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
