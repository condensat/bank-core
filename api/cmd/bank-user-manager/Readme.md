# BankUserManagement

This tool imports user and credential from file into database
Logins and passwords are hashed before storing.

New users are created.
Credentials are override for existing user (depending on `hash_seed` flag or `$PasswordHashSeed` env)

## Usage

From file:

```bash
	go run go run api/cmd/bank-user-manager/main.go --userFile=<userfile> 
```

From Stdin:

```bash
	go run go run api/cmd/bank-user-manager/main.go < <userfile>
```

## Userfile format

````
<login>:<password>:<valid_email>:roles[](admin,user)
```

## Roles

Roles are liste of named roles allowing access control to ressources.