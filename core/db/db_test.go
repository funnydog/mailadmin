package db

import (
	"os"
	"testing"

	"github.com/funnydog/mailadmin/core/config"
)

func TestSQLite3(t *testing.T) {
	conf := config.Configuration{
		DBType: "sqlite3",
		DBName: "/tmp/test.sqlite",
	}
	db, err := Connect(&conf)
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(conf.DBName)
	defer db.Close()

	_, err = db.Db.Exec("CREATE TABLE testusers(id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Error(err)
	}

	err = db.PrepareStatement("users", "SELECT * FROM testusers")
	if err != nil {
		t.Error(err)
	}

	_, _ = db.Db.Exec("DROP TABLE testusers")
}

func TestPostgreSQL(t *testing.T) {
	conf := config.Configuration{
		DBType:    "postgres",
		DBUser:    "postgres",
		DBName:    "django",
		DBSSLMode: "disable",
	}

	db, err := Connect(&conf)
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()

	_, err = db.Db.Exec("CREATE TABLE testusers(id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Error(err)
	}

	err = db.PrepareStatement("users", "SELECT * FROM testusers")
	if err != nil {
		t.Error(err)
	}

	_, _ = db.Db.Exec("DROP TABLE testusers")
}

func TestUnknown(t *testing.T) {
	conf := config.Configuration{
		DBType: "unknown",
	}

	db, err := Connect(&conf)
	if err == nil {
		t.Error("Connected to unknown DBType")
		db.Close()
	}
}
