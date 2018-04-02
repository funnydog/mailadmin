package types

import (
	"database/sql"
	"time"

	"github.com/funnydog/mailadmin/core/db"
)

type Domain struct {
	Id          sql.NullInt64
	Name        string
	Description string
	BackupMX    bool
	Active      bool
	Created     time.Time
	Modified    time.Time
}

func (domain *Domain) Create(db *db.Database) error {
	stmt, err := db.FindStatement("domainCreate")
	if err != nil {
		return err
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

func (domain *Domain) Update(db *db.Database) error {
	stmt, err := db.FindStatement("domainUpdate")
	if err != nil {
		return err
	}

	domain.Modified = time.Now()

	_, err = stmt.Exec(
		domain.Name,
		domain.Description,
		domain.BackupMX,
		domain.Active,
		domain.Modified,
		domain.Id,
	)

	return err
}

func (domain *Domain) Delete(db *db.Database) error {
	stmt, err := db.FindStatement("domainDelete")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(domain.Id.Int64)
	return err
}

func GetDomainList(db *db.Database) ([]Domain, error) {
	domains := []Domain{}

	stmt, err := db.FindStatement("domainList")
	if err != nil {
		return domains, err
	}

	rows, err := stmt.Query()
	if err != nil {
		return domains, err
	}
	defer rows.Close()

	for rows.Next() {
		t := Domain{}
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

func GetDomain(db *db.Database, PK int64) (Domain, error) {
	t := Domain{}

	stmt, err := db.FindStatement("domainFind")
	if err != nil {
		return t, err
	}

	err = stmt.QueryRow(PK).Scan(
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

type Mailbox struct {
	Id       sql.NullInt64
	Domain   sql.NullInt64
	Email    string
	Password string
	Created  time.Time
	Modified time.Time
	Active   bool
}

func (mailbox *Mailbox) Create(db *db.Database) error {
	stmt, err := db.FindStatement("mailboxCreate")
	if err != nil {
		return err
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

func (mailbox *Mailbox) Update(db *db.Database) error {
	stmt, err := db.FindStatement("mailboxUpdate")
	if err != nil {
		return err
	}

	mailbox.Modified = time.Now()

	_, err = stmt.Exec(
		mailbox.Domain,
		mailbox.Email,
		mailbox.Password,
		mailbox.Active,
		mailbox.Modified,
		mailbox.Id,
	)

	return err
}

func (mailbox Mailbox) Delete(db *db.Database) error {
	stmt, err := db.FindStatement("mailboxDelete")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(mailbox.Id.Int64)
	return err
}

func GetMailboxList(db *db.Database, domain_id int64) ([]Mailbox, error) {
	mailboxes := []Mailbox{}

	stmt, err := db.FindStatement("mailboxList")
	if err != nil {
		return mailboxes, err
	}

	rows, err := stmt.Query(domain_id)
	if err != nil {
		return mailboxes, err
	}
	defer rows.Close()

	for rows.Next() {
		t := Mailbox{}
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

func GetMailbox(db *db.Database, PK int64) (Mailbox, error) {
	var t Mailbox

	stmt, err := db.FindStatement("mailboxFind")
	if err != nil {
		return t, err
	}

	err = stmt.QueryRow(PK).Scan(
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

type Alias struct {
	Id          sql.NullInt64
	Domain      sql.NullInt64
	Source      string
	Destination string
	Created     time.Time
	Modified    time.Time
	Active      bool
}

func (alias *Alias) Create(db *db.Database) error {
	stmt, err := db.FindStatement("aliasCreate")
	if err != nil {
		return err
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

func (alias *Alias) Update(db *db.Database) error {
	stmt, err := db.FindStatement("aliasUpdate")
	if err != nil {
		return err
	}

	alias.Modified = time.Now()

	_, err = stmt.Exec(
		alias.Domain,
		alias.Source,
		alias.Destination,
		alias.Active,
		alias.Modified,
		alias.Id,
	)
	return err
}

func (alias Alias) Delete(db *db.Database) error {
	stmt, err := db.FindStatement("aliasDelete")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(alias.Id.Int64)
	return err
}

func GetAliasList(db *db.Database, domain_id int64) ([]Alias, error) {
	aliases := []Alias{}

	stmt, err := db.FindStatement("aliasList")
	if err != nil {
		return aliases, err
	}

	rows, err := stmt.Query(domain_id)
	if err != nil {
		return aliases, err
	}
	defer rows.Close()

	for rows.Next() {
		t := Alias{}
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

func GetAlias(db *db.Database, PK int64) (Alias, error) {
	t := Alias{}

	stmt, err := db.FindStatement("aliasFind")
	if err != nil {
		return t, err
	}

	err = stmt.QueryRow(PK).Scan(
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

func RegisterDatabase(db *db.Database) error {
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

	for key, sql := range stmts {
		err := db.PrepareStatement(key, sql)
		if err != nil {
			return err
		}
	}
	return nil
}
