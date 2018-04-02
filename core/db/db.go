package db

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

var (
	statementNotFound        = errors.New("Statement not found in the statements map[]")
	statementAlreadyInserted = errors.New("Statement already inserted in the statements map[]")
)

type Database struct {
	db    *sql.DB
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

	stmt, err := db.db.Prepare(sql)
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
	db.db.Close()
}

func Connect(conn string) (Database, error) {
	db, err := sql.Open("sqlite3", conn)
	if err != nil {
		return Database{}, err
	}

	return Database{
		db:    db,
		stmts: map[string]*sql.Stmt{},
	}, nil
}
