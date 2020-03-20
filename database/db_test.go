// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/logger"

	"github.com/jinzhu/gorm"
)

func setup(ctx context.Context, databaseName string, models []model.Model) context.Context {
	options := Options{
		HostName:      "localhost",
		Port:          3306,
		User:          "condensat",
		Password:      "condensat",
		Database:      "condensat",
		EnableLogging: false,
	}
	if databaseName == options.Database {
		panic("Wrong databaseName")
	}

	ctx = appcontext.WithDatabase(ctx, NewDatabase(options))
	db := appcontext.Database(ctx).DB().(*gorm.DB)

	createDatabase := fmt.Sprintf("create database if not exists %s; use %s;", databaseName, databaseName)
	db.Exec(createDatabase)

	err := db.Exec(createDatabase).Error
	if err != nil {
		panic(err)
	}

	migrateDatabase(ctx, models)

	return ctx
}

func teardown(ctx context.Context, databaseName string) {
	db := appcontext.Database(ctx).DB().(*gorm.DB)

	dropDatabase := fmt.Sprintf("drop database if exists %s", databaseName)
	err := db.Exec(dropDatabase).Error
	if err != nil {
		panic(err)
	}
}

func migrateDatabase(ctx context.Context, models []model.Model) {
	db := appcontext.Database(ctx)

	err := db.Migrate(models)
	if err != nil {
		logger.Logger(ctx).WithError(err).
			WithField("Method", "main.migrateDatabase").
			Panic("Failed to migrate database models")
	}
}

func getSortedTypeFileds(t reflect.Type) []string {
	count := t.NumField()
	result := make([]string, 0, count)

	for i := 0; i < count; i++ {
		field := gorm.TheNamingStrategy.Column(t.Field(i).Name)
		result = append(result, field)
	}

	for i, field := range result {
		result[i] = gorm.TheNamingStrategy.Column(field)
	}
	sort.Strings(result)

	return result
}
