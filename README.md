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

## Build a static executable

The command ```go build``` will build a single executable dynamically
linked with the various libraries (included the libc).

That executable is not portable to all the Linux distributions because
of a mismatch of the libc version.

You can statically link the executable but not against the GNU libc
(glibc). You can use the musl libc instead.

Provided that the musl libc is installed the following command will
build a complete statically linked executable:

```
$ CGO_CFLAGS="-D_LARGEFILE64_SOURCE" CC=musl-gcc go build --ldflags '-linkmode external -extldflags "-static"'
```

## Disclaimer

May contain unintended bugs.
