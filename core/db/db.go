package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/funnydog/mailadmin/core/config"

	_ "github.com/mattn/go-sqlite3"
)

var (
	databaseTypeNotSupported = errors.New("Database type not supported")
	statementNotFound        = errors.New("Statement not found in the statements map[]")
	statementAlreadyInserted = errors.New("Statement already inserted in the statements map[]")
)

type Database struct {
	Db    *sql.DB
	stmts map[string]*sql.Stmt
}

func (db *Database) FindStatement(key string) (*sql.Stmt, error) {
	stmt, ok := db.stmts[key]
	if !ok {
		return nil, statementNotFound
	}
	return stmt, nil
}

func (db *Database) PrepareStatement(key, sql string) error {
	_, ok := db.stmts[key]
	if ok {
		return statementAlreadyInserted
	}

	stmt, err := db.Db.Prepare(sql)
	if err != nil {
		return err
	}

	db.stmts[key] = stmt
	return nil
}

func (db *Database) Close() {
	for _, stmt := range db.stmts {
		stmt.Close()
	}
	db.Db.Close()
}

func Connect(conf *config.Configuration) (*Database, error) {
	var (
		db  *sql.DB
		err error
	)
	switch conf.DBType {
	case "sqlite3":
		db, err = sql.Open("sqlite3", conf.DBName)
	case "postgres":
		parameters := []string{}
		if conf.DBUser != "" {
			parameters = append(parameters, fmt.Sprintf("user=%s", conf.DBUser))
		}

		if conf.DBPass != "" {
			parameters = append(parameters, fmt.Sprintf("password=%s", conf.DBPass))
		}

		if conf.DBName != "" {
			parameters = append(parameters, fmt.Sprintf("dbname=%s", conf.DBName))
		}

		if conf.DBHost != "" {
			parameters = append(parameters, fmt.Sprintf("host=%s", conf.DBHost))
		}

		if conf.DBPort != "" {
			parameters = append(parameters, fmt.Sprintf("port=%s", conf.DBPort))
		}

		if conf.DBSSLMode != "" {
			parameters = append(parameters, fmt.Sprintf("sslmode=%s", conf.DBSSLMode))
		}

		db, err = sql.Open("postgres", strings.Join(parameters, " "))
	default:
		err = databaseTypeNotSupported
	}

	if err != nil {
		return nil, err
	}

	return &Database{
		Db:    db,
		stmts: map[string]*sql.Stmt{},
	}, nil
}
