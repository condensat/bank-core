// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"testing"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/database/model"
	"github.com/jinzhu/gorm"
)

func TestUserHasRole(t *testing.T) {
	const databaseName = "TestUserHasRole"
	t.Parallel()

	db := setup(databaseName, UserModel())
	defer teardown(db, databaseName)

	user, err := FindOrCreateUser(db, model.User{
		Name:  "test",
		Email: "test@condensat.tech",
	})
	if err != nil {
		t.Errorf("Unable to add user")
		return
	}
	// check 1..N relation
	addUserRole(db, user.ID, model.RoleName("foo"))
	addUserRole(db, user.ID, model.RoleName("bar"))

	admin, err := FindOrCreateUser(db, model.User{
		Name:  "admin",
		Email: "admin@condensat.tech",
	})
	if err != nil {
		t.Errorf("Unable to add admin")
		return
	}
	// check 1..N relation for multiple users
	addUserRole(db, admin.ID, model.RoleName("foo"))
	addUserRole(db, admin.ID, model.RoleName("bar"))
	addUserRole(db, admin.ID, model.RoleNameAdmin)

	type args struct {
		userID model.UserID
		role   model.RoleName
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"default", args{}, false, true},

		{"valid", args{user.ID, model.RoleNameDefault}, true, false},

		{"foo", args{user.ID, model.RoleName("foo")}, true, false},
		{"bar", args{user.ID, model.RoleName("bar")}, true, false},
		{"foobar", args{user.ID, model.RoleName("foobar")}, false, false},

		{"user", args{user.ID, model.RoleNameAdmin}, false, false},
		{"admin", args{admin.ID, model.RoleNameAdmin}, true, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := UserHasRole(db, tt.args.userID, tt.args.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserHasRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UserHasRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func addUserRole(db bank.Database, userID model.UserID, role model.RoleName) {
	gdb := db.DB().(*gorm.DB)

	err := gdb.Create(&model.UserRole{
		UserID: userID,
		Role:   role,
	}).Error
	if err != nil {
		panic(err)
	}
}
