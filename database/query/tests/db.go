// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tests

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"

	"github.com/condensat/bank-core/database"

	"github.com/jinzhu/gorm"
)

func Setup(databaseName string, models []database.Model) database.Context {
	options := database.Options{
		HostName:      "mariadb",
		Port:          3306,
		User:          "condensat",
		Password:      "condensat",
		Database:      "condensat",
		EnableLogging: false,
	}
	if databaseName == options.Database {
		panic("Wrong databaseName")
	}

	db := database.New(options)
	gdb := db.DB().(*gorm.DB)

	createDatabase := fmt.Sprintf("create database if not exists %s; use %s;", databaseName, databaseName)
	gdb.Exec(createDatabase)

	err := gdb.Exec(createDatabase).Error
	if err != nil {
		panic(err)
	}

	MigrateDatabase(db, models)

	return db
}

func Teardown(db database.Context, databaseName string) {
	gdb := db.DB().(*gorm.DB)

	dropDatabase := fmt.Sprintf("drop database if exists %s", databaseName)
	err := gdb.Exec(dropDatabase).Error
	if err != nil {
		panic(err)
	}
}

func MigrateDatabase(db database.Context, models []database.Model) {
	err := db.Migrate(models)
	if err != nil {
		panic(err)
	}
}

func GetSortedTypeFileds(t reflect.Type) []string {
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

var (
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func RandSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
