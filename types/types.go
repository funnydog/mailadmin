package types

import (
	"database/sql"
	"time"
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

type Mailbox struct {
	Id       sql.NullInt64
	Domain   sql.NullInt64
	Email    string
	Password string
	Created  time.Time
	Modified time.Time
	Active   bool
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
