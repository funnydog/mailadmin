package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/funnydog/mailadmin/core/config"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type ErrDbTypeNotSupported string

func (db ErrDbTypeNotSupported) Error() string {
	return fmt.Sprintf("Database type '%s' is not supported", string(db))
}

type ErrStatementAlreadyPresent string

func (sp ErrStatementAlreadyPresent) Error() string {
	return fmt.Sprintf("Statement '%s' already present in the map", string(sp))
}

type ErrStatementNotFound string

func (sn ErrStatementNotFound) Error() string {
	return fmt.Sprintf("Statement '%s' not found in the map", string(sn))
}

type Database struct {
	Db    *sql.DB
	stmts map[string]*sql.Stmt
}

func (db *Database) PrepareStatement(key, sql string) error {
	_, ok := db.stmts[key]
	if ok {
		return ErrStatementAlreadyPresent(key)
	}

	stmt, err := db.Db.Prepare(sql)
	if err != nil {
		return err
	}

	db.stmts[key] = stmt
	return nil
}

func (db *Database) FindStatement(key string) (*sql.Stmt, error) {
	stmt, ok := db.stmts[key]
	if !ok {
		return nil, ErrStatementNotFound(key)
	}
	return stmt, nil
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
		if err != nil {
			return nil, err
		}

		_, err = db.Exec("PRAGMA foreign_keys = ON;")
		if err != nil {
			return nil, err
		}

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
		if err != nil {
			return nil, err
		}

	default:
		return nil, ErrDbTypeNotSupported(conf.DBType)
	}

	return &Database{
		Db:    db,
		stmts: map[string]*sql.Stmt{},
	}, nil
}
