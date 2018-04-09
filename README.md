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
   $ go get -u github.com/julienschmidt/httprouter
   $ go get -u github.com/mattn/go-sqlite3
   $ go get -u github.com/pborman/getopt/v2
   ```

2. Create the sqlite3 database by invoking the application with the -m
   flag: ```go run mailadmin.go -m``` The database will be named as
   the dbname field in config.json.

3. Edit config.json to change the password for signing in the web
   application. You can use for example ```doveadm pw -s
   SHA512-CRYPT``` to generate a new hash. You can reuse the current
   hash which matches the password ```pass```.

4. Run the application: ```go run mailadmin.go``` and connect to
   localhost:8080 to sign-in. The default username is ```admin``` and
   the default password is ```pass```. You can also compile the
   project with ```go build mailadmin.go``` and run the resulting
   binary.

## Disclaimer

May contain unintended bugs.
