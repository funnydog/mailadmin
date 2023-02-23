## MailAdmin

MailAdmin is a simple web application written in golang for managing
the virtual domains for a mail server.

The data is stored in a sqlite database which is then used by dovecot
and postfix to get the appropriate data.

The application stores the passwords hashed with the bcrypt algorithm
with a cost of 10.

## Motivation

This application was created to ease the management of my personal
mail server.

## Quick start

1. Get the necessary dependencies:

   ```
   $ go mod tidy
   ```

2. Change the config.json file appropriately. The password is the
   bcrypt hash of the actual password for signing in. To change it
   just invoke the application from the command line with the
   following command: ```go run mailadmin.go -p```. You can reuse the
   current hash which matches the password ```pass```.

3. Create the sqlite3 database by invoking the application with the -m
   flag: ```go run mailadmin.go -m``` The database will be named as
   the dbname field in config.json.

4. Run the application: ```go run mailadmin.go``` and connect to
   localhost:8080 to sign-in. The default username is ```admin``` and
   the default password is ```pass```. You can also compile the
   project with ```go build mailadmin.go``` and run the resulting
   binary.

## Disclaimer

May contain unintended bugs.
