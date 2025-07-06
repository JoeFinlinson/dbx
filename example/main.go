package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/JoeFinlinson/dbx"
	"github.com/jackc/pgx/v5/pgxpool"
)

// User represents a user in the database
// Using the explicit table.column naming convention
type User struct {
	Users_ID     int    `db:"users.id"`
	Users_Name   string `db:"users.name"`
	Users_Email  string `db:"users.email"`
	Users_Active bool   `db:"users.active"`
}

// Invoice represents an invoice with customer information
// Demonstrates joining multiple tables
type Invoice struct {
	Invoice_ID     int     `db:"invoice.id"`
	Invoice_Amount float64 `db:"invoice.amount"`
	Customer_Email string  `db:"customer.email"`
	Customer_Name  string  `db:"customer.name"`
}

func main() {
	ctx := context.Background()

	// Connect to PostgreSQL
	// Note: You'll need to update this connection string for your setup
	dbpool, err := pgxpool.New(ctx, "postgres://username:password@localhost:5432/dbx_example?sslmode=disable")
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		log.Println("Please update the connection string in example/main.go")
		return
	}
	defer dbpool.Close()

	// Create a dbx.DB interface from pgxpool
	// pgxpool.Pool implements the dbx.DB interface
	var db dbx.DB = dbpool

	// Example 1: QueryMaps - Get results as maps (no struct needed)
	fmt.Println("=== QueryMaps Example ===")
	rows, err := dbx.QueryMaps(ctx, db, "SELECT * FROM users WHERE active = $1", true)
	if err != nil {
		log.Printf("QueryMaps failed: %v", err)
	} else {
		for i, row := range rows {
			fmt.Printf("User %d: %+v\n", i+1, row)
		}
	}

	// Example 2: QueryJSON - Get results as JSON
	fmt.Println("\n=== QueryJSON Example ===")
	jsonData, err := dbx.QueryJSON(ctx, db, "SELECT name, email FROM users LIMIT 3")
	if err != nil {
		log.Printf("QueryJSON failed: %v", err)
	} else {
		fmt.Printf("JSON result: %s\n", string(jsonData))
	}

	// Example 3: InsertStruct - Insert a struct into a table
	fmt.Println("\n=== InsertStruct Example ===")
	user := User{
		Users_Name:   "Alice Johnson",
		Users_Email:  "alice@example.com",
		Users_Active: true,
	}

	err = dbx.InsertStruct(ctx, db, "users", user)
	if err != nil {
		log.Printf("InsertStruct failed: %v", err)
	} else {
		fmt.Println("Successfully inserted user:", user.Users_Name)
	}

	// Example 4: QueryStructs - Query into struct slice
	fmt.Println("\n=== QueryStructs Example ===")
	var users []User
	err = dbx.QueryStructs(ctx, db, "SELECT * FROM users ORDER BY id LIMIT 5", &users)
	if err != nil {
		log.Printf("QueryStructs failed: %v", err)
	} else {
		for _, u := range users {
			fmt.Printf("User: ID=%d, Name=%s, Email=%s, Active=%t\n",
				u.Users_ID, u.Users_Name, u.Users_Email, u.Users_Active)
		}
	}

	// Example 5: Complex join with explicit table.column mapping
	fmt.Println("\n=== Complex Join Example ===")
	var invoices []Invoice
	err = dbx.QueryStructs(ctx, db, `
		SELECT 
			i.id as "invoice.id",
			i.amount as "invoice.amount",
			c.email as "customer.email",
			c.name as "customer.name"
		FROM invoice i
		JOIN customer c ON i.customer_id = c.id
		WHERE i.amount > $1
		ORDER BY i.amount DESC
		LIMIT 3
	`, &invoices, 100.0)
	if err != nil {
		log.Printf("Complex join failed: %v", err)
	} else {
		for _, inv := range invoices {
			fmt.Printf("Invoice: ID=%d, Amount=$%.2f, Customer=%s (%s)\n",
				inv.Invoice_ID, inv.Invoice_Amount, inv.Customer_Name, inv.Customer_Email)
		}
	}

	// Example 6: Dynamic query with maps
	fmt.Println("\n=== Dynamic Query Example ===")
	// Simulate a dynamic filter
	filterActive := true
	filterEmail := "%@example.com"

	var query string
	var args []any

	if filterActive {
		query = "SELECT * FROM users WHERE active = $1"
		args = append(args, filterActive)
	} else {
		query = "SELECT * FROM users"
	}

	if filterEmail != "" {
		if len(args) > 0 {
			query += " AND email LIKE $2"
		} else {
			query += " WHERE email LIKE $1"
		}
		args = append(args, filterEmail)
	}

	query += " LIMIT 10"

	dynamicRows, err := dbx.QueryMaps(ctx, db, query, args...)
	if err != nil {
		log.Printf("Dynamic query failed: %v", err)
	} else {
		fmt.Printf("Dynamic query returned %d rows\n", len(dynamicRows))
		for _, row := range dynamicRows {
			fmt.Printf("  - %s (%s)\n", row["name"], row["email"])
		}
	}
}

// Helper function to pretty print JSON
func prettyPrintJSON(data []byte) {
	var prettyJSON interface{}
	if err := json.Unmarshal(data, &prettyJSON); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return
	}

	prettyData, err := json.MarshalIndent(prettyJSON, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting JSON: %v\n", err)
		return
	}

	fmt.Println(string(prettyData))
}
