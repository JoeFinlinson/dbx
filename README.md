# dbx – SQL-first Data Access for Go

A lightweight, opinionated data access layer for Go that keeps SQL front and center while eliminating boilerplate.

## Philosophy

- ✅ **Raw SQL** - You write the SQL, we handle the mapping
- ✅ **No ORM** - No magic, no hidden queries, no abstraction layers
- ✅ **Explicit mapping** - Use `db:"table.column"` tags for clarity
- ✅ **Flexible output** - Get maps, structs, or JSON as needed
- ✅ **Team-friendly** - Code that's obvious to any Go developer

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/jfinlinson/dbx"
)

type User struct {
    Users_ID    int    `db:"users.id"`
    Users_Name  string `db:"users.name"`
    Users_Email string `db:"users.email"`
}

func main() {
    ctx := context.Background()
    dbpool, err := pgxpool.New(ctx, "postgres://user:pass@localhost:5432/dbname")
    if err != nil {
        log.Fatal(err)
    }
    defer dbpool.Close()

    // Query into maps (no struct needed)
    rows, err := dbx.QueryMaps(ctx, dbpool, "SELECT * FROM users WHERE email LIKE $1", "%@example.com")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Users: %+v\n", rows)

    // Query into structs
    var users []User
    err = dbx.QueryStructs(ctx, dbpool, "SELECT * FROM users", &users)
    if err != nil {
        log.Fatal(err)
    }

    // Insert a struct
    user := User{Users_Name: "Alice", Users_Email: "alice@example.com"}
    err = dbx.InsertStruct(ctx, dbpool, "users", user)
    if err != nil {
        log.Fatal(err)
    }
}
```

## Features

### QueryMaps
Get query results as `[]map[string]interface{}` - perfect for dynamic data, APIs, or when you don't want to define structs.

```go
rows, err := dbx.QueryMaps(ctx, db, "SELECT * FROM users WHERE active = $1", true)
```

### QueryStructs
Map query results into structs using `db:"table.column"` tags for explicit mapping.

```go
type Invoice struct {
    Invoice_Amount  float64 `db:"invoice.amount"`
    Customer_Email  string  `db:"customer.email"`
}

var invoices []Invoice
err := dbx.QueryStructs(ctx, db, "SELECT invoice.amount, customer.email FROM invoice JOIN customer ON ...", &invoices)
```

### InsertStruct
Insert a struct into a table, automatically mapping fields to columns.

```go
user := User{Users_Name: "Bob", Users_Email: "bob@example.com"}
err := dbx.InsertStruct(ctx, db, "users", user)
```

### QueryJSON
Get query results as JSON bytes - great for APIs.

```go
jsonData, err := dbx.QueryJSON(ctx, db, "SELECT * FROM users")
```

## Design Principles

1. **SQL First** - You write SQL, we handle the rest
2. **Explicit Mapping** - Use `db:"table.column"` tags to make relationships clear
3. **No Magic** - Everything is visible and debuggable
4. **Performance** - Built on pgx for maximum PostgreSQL performance
5. **Team Clarity** - Code that any Go developer can understand immediately

## Installation

```bash
go get github.com/JoeFinlinson/dbx
```

## License

MIT License - see LICENSE file for details. 
