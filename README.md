## MailAdmin

MailAdmin is a simple web application written in golang for managing
the virtual domains for a mail server.

## Motivation

Easen the management of my mail server.

## Quick start

1. Get the necessary dependencies:

   ```
   $ go get -u github.com/go-errors/errors
   $ go get -u github.com/gorilla/csrf
   $ go get -u github.com/gorilla/sessions
   $ go get -u github.com/julienshmidt/httprouter
   $ go get -u github.com/mattn/go-sqlite3
   ```

2. Create the sqlite3 database using the tables.sql schema:
   ```$ sqlite3 postfix.db < tables.sql```.

3. Edit the config.json file to correctly point to the db.  Here you
   can also change the password for signing in the web
   application. You can use for example ```doveadm pw -s
   SHA512-CRYPT``` to generate a new hash.

4. Run the application: ```go run mailadmin.go``` and connect to
   localhost:8080 to sign-in. The default password is ```pass```. You
   can also compile the project with ```go build mailadmin.go``` and
   run the resulting binary.

## Disclaimer

May contain unintended bugs.
