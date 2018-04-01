package db

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/funnydog/mailadmin/types"
)

var queryNotFound = errors.New("Query not found in the statement map[]")

type Database struct {
	db    *sql.DB
	stmts map[string]*sql.Stmt
}

func (db *Database) GetDomainList() ([]types.Domain, error) {
	domains := []types.Domain{}

	stmt, ok := db.stmts["domainList"]
	if !ok {
		return domains, queryNotFound
	}

	rows, err := stmt.Query()
	if err != nil {
		return domains, err
	}
	defer rows.Close()

	for rows.Next() {
		var t types.Domain
		err := rows.Scan(
			&t.Id,
			&t.Name,
			&t.Description,
			&t.BackupMX,
			&t.Active,
			&t.Created,
			&t.Modified,
		)
		if err != nil {
			return domains, err
		}

		domains = append(domains, t)
	}
	return domains, rows.Err()
}

func (db *Database) GetDomain(PK int64) (types.Domain, error) {
	t := types.Domain{}

	stmt, ok := db.stmts["domainFind"]
	if !ok {
		return t, queryNotFound
	}

	err := stmt.QueryRow(PK).Scan(
		&t.Id,
		&t.Name,
		&t.Description,
		&t.BackupMX,
		&t.Active,
		&t.Created,
		&t.Modified,
	)
	return t, err
}

func (db *Database) CreateDomain(domain *types.Domain) error {
	stmt, ok := db.stmts["domainCreate"]
	if !ok {
		return queryNotFound
	}

	domain.Created = time.Now()
	domain.Modified = domain.Created

	res, err := stmt.Exec(
		domain.Name,
		domain.Description,
		domain.BackupMX,
		domain.Active,
		domain.Created,
		domain.Modified,
	)
	if err != nil {
		return err
	}

	domain.Id.Int64, err = res.LastInsertId()
	if err != nil {
		return err
	}
	domain.Id.Valid = true
	return nil
}

func (db *Database) UpdateDomain(domain *types.Domain) error {
	stmt, ok := db.stmts["domainUpdate"]
	if !ok {
		return queryNotFound
	}

	domain.Modified = time.Now()

	_, err := stmt.Exec(
		domain.Name,
		domain.Description,
		domain.BackupMX,
		domain.Active,
		domain.Modified,
		domain.Id,
	)

	return err
}

func (db *Database) DeleteDomain(PK int64) error {
	stmt, ok := db.stmts["domainDelete"]
	if !ok {
		return queryNotFound
	}

	_, err := stmt.Exec(PK)
	return err
}

func (db *Database) GetMailboxList(domain_id int64) ([]types.Mailbox, error) {
	mailboxes := []types.Mailbox{}

	stmt, ok := db.stmts["mailboxList"]
	if !ok {
		return mailboxes, queryNotFound
	}

	rows, err := stmt.Query(domain_id)
	if err != nil {
		return mailboxes, err
	}
	defer rows.Close()

	for rows.Next() {
		var t types.Mailbox
		err := rows.Scan(
			&t.Id,
			&t.Domain,
			&t.Email,
			&t.Password,
			&t.Active,
			&t.Created,
			&t.Modified,
		)
		if err != nil {
			return mailboxes, err
		}

		mailboxes = append(mailboxes, t)
	}
	return mailboxes, rows.Err()
}

func (db *Database) GetMailbox(PK int64) (types.Mailbox, error) {
	var t types.Mailbox

	stmt, ok := db.stmts["mailboxFind"]
	if !ok {
		return t, queryNotFound
	}

	err := stmt.QueryRow(PK).Scan(
		&t.Id,
		&t.Domain,
		&t.Email,
		&t.Password,
		&t.Active,
		&t.Created,
		&t.Modified,
	)
	return t, err
}

func (db *Database) CreateMailbox(mailbox *types.Mailbox) error {
	stmt, ok := db.stmts["mailboxCreate"]
	if !ok {
		return queryNotFound
	}

	mailbox.Created = time.Now()
	mailbox.Modified = mailbox.Created

	res, err := stmt.Exec(
		mailbox.Domain,
		mailbox.Email,
		mailbox.Password,
		mailbox.Active,
		mailbox.Created,
		mailbox.Modified,
	)
	if err != nil {
		return err
	}

	mailbox.Id.Int64, err = res.LastInsertId()
	if err != nil {
		return err
	}
	mailbox.Id.Valid = true
	return nil
}

func (db *Database) UpdateMailbox(mailbox *types.Mailbox) error {
	stmt, ok := db.stmts["mailboxUpdate"]
	if !ok {
		return queryNotFound
	}

	mailbox.Modified = time.Now()

	_, err := stmt.Exec(
		mailbox.Domain,
		mailbox.Email,
		mailbox.Password,
		mailbox.Active,
		mailbox.Modified,
		mailbox.Id,
	)

	return err
}

func (db *Database) DeleteMailbox(PK int64) error {
	stmt, ok := db.stmts["mailboxDelete"]
	if !ok {
		return queryNotFound
	}

	_, err := stmt.Exec(PK)
	return err
}

func (db *Database) GetAliasList(domain_id int64) ([]types.Alias, error) {
	aliases := []types.Alias{}

	stmt, ok := db.stmts["aliasList"]
	if !ok {
		return aliases, queryNotFound
	}

	rows, err := stmt.Query(domain_id)
	if err != nil {
		return aliases, err
	}
	defer rows.Close()

	for rows.Next() {
		var t types.Alias
		err := rows.Scan(
			&t.Id,
			&t.Domain,
			&t.Source,
			&t.Destination,
			&t.Active,
			&t.Created,
			&t.Modified,
		)
		if err != nil {
			return aliases, err
		}

		aliases = append(aliases, t)
	}
	return aliases, rows.Err()
}

func (db *Database) GetAlias(PK int64) (types.Alias, error) {
	var t types.Alias

	stmt, ok := db.stmts["aliasFind"]
	if !ok {
		return t, queryNotFound
	}

	err := stmt.QueryRow(PK).Scan(
		&t.Id,
		&t.Domain,
		&t.Source,
		&t.Destination,
		&t.Active,
		&t.Created,
		&t.Modified,
	)
	return t, err
}

func (db *Database) CreateAlias(alias *types.Alias) error {
	stmt, ok := db.stmts["aliasCreate"]
	if !ok {
		return queryNotFound
	}

	alias.Created = time.Now()
	alias.Modified = alias.Created

	res, err := stmt.Exec(
		alias.Domain,
		alias.Source,
		alias.Destination,
		alias.Active,
		alias.Created,
		alias.Modified,
	)
	if err != nil {
		return err
	}

	alias.Id.Int64, err = res.LastInsertId()
	if err != nil {
		return err
	}
	alias.Id.Valid = true
	return nil
}

func (db *Database) UpdateAlias(alias *types.Alias) error {
	stmt, ok := db.stmts["aliasUpdate"]
	if !ok {
		return queryNotFound
	}

	alias.Modified = time.Now()

	_, err := stmt.Exec(
		alias.Domain,
		alias.Source,
		alias.Destination,
		alias.Active,
		alias.Modified,
		alias.Id,
	)
	return err
}

func (db *Database) DeleteAlias(PK int64) error {
	stmt, ok := db.stmts["aliasDelete"]
	if !ok {
		return queryNotFound
	}

	_, err := stmt.Exec(PK)
	return err
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

	stmts := map[string]string{
		// domains
		"domainList":   `SELECT id, name, description, backupmx, active, created, modified FROM domain ORDER BY name`,
		"domainFind":   `SELECT id, name, description, backupmx, active, created, modified FROM domain WHERE id=$1`,
		"domainCreate": `INSERT INTO domain(name, description, backupmx, active, created, modified) VALUES ($1, $2, $3, $4, $5, $6)`,
		"domainUpdate": `UPDATE domain SET name=$1, description=$2, backupmx=$3, active=$4, modified=$5 WHERE id=$7`,
		"domainDelete": `DELETE FROM domain WHERE id=$1`,

		// mailboxes
		"mailboxList":   `SELECT id, domain_id, email, password, active, created, modified FROM mailbox WHERE domain_id=$1 ORDER BY email`,
		"mailboxFind":   `SELECT id, domain_id, email, password, active, created, modified FROM mailbox WHERE id=$1`,
		"mailboxCreate": `INSERT INTO mailbox(domain_id, email, password, active, created, modified) VALUES ($1, $2, $3, $4, $5, $6)`,
		"mailboxUpdate": `UPDATE mailbox SET domain_id=$1, email=$2, password=$3, active=$4, modified=$5 WHERE id=$6`,
		"mailboxDelete": `DELETE FROM mailbox WHERE id=$1`,

		// aliases
		"aliasList":   `SELECT id, domain_id, source, destination, active, created, modified FROM alias WHERE domain_id=$1 ORDER BY source`,
		"aliasFind":   `SELECT id, domain_id, source, destination, active, created, modified FROM alias WHERE id=$1`,
		"aliasCreate": `INSERT INTO alias(domain_id, source, destination, active, created, modified) VALUES ($1, $2, $3, $4, $5, $6)`,
		"aliasUpdate": `UPDATE alias SET domain_id=$1, source=$2, destination=$3, active=$4, modified=$5 WHERE id=$6`,
		"aliasDelete": `DELETE FROM alias WHERE id=$1`,
	}

	prepared := map[string]*sql.Stmt{}
	for key, value := range stmts {
		stmt, err := db.Prepare(value)
		if err != nil {
			for _, stmt := range prepared {
				stmt.Close()
			}
			return Database{}, err
		}
		prepared[key] = stmt
	}
	return Database{
		db:    db,
		stmts: prepared,
	}, nil
}
