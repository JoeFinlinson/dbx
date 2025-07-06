// Package dbx provides SQL-first data access helpers for Go, built on top of pgx.
// It eliminates boilerplate while keeping SQL front and center.
//
// Key features:
//   - QueryMaps: Get results as []map[string]interface{}
//   - QueryStructs: Map results into structs using db:"table.column" tags
//   - InsertStruct: Insert structs into tables automatically
//   - QueryJSON: Get results as JSON bytes
//
// Example:
//
//	rows, err := dbx.QueryMaps(ctx, db, "SELECT * FROM users WHERE active = $1", true)
//	var users []User
//	err = dbx.QueryStructs(ctx, db, "SELECT * FROM users", &users)
//	err = dbx.InsertStruct(ctx, db, "users", user)
package dbx

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DB is an interface that provides the methods needed for database operations
type DB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// RowMap represents a single database row as a map
type RowMap map[string]interface{}

// QueryMaps executes a query and returns results as a slice of maps.
// Each map represents a row with column names as keys.
func QueryMaps(ctx context.Context, db DB, sql string, args ...any) ([]RowMap, error) {
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	fieldDescs := rows.FieldDescriptions()
	fieldNames := make([]string, len(fieldDescs))
	for i, fd := range fieldDescs {
		fieldNames[i] = string(fd.Name)
	}

	var result []RowMap
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to get row values: %w", err)
		}

		row := make(RowMap, len(values))
		for i, v := range values {
			row[fieldNames[i]] = v
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return result, nil
}

// QueryJSON executes a query and returns results as JSON bytes.
// This is useful for APIs or when you need JSON output directly.
func QueryJSON(ctx context.Context, db DB, sql string, args ...any) ([]byte, error) {
	rows, err := QueryMaps(ctx, db, sql, args...)
	if err != nil {
		return nil, err
	}
	return json.Marshal(rows)
}

// InsertStruct inserts a struct into the specified table.
// It uses db:"column" tags to map struct fields to table columns.
// Fields without db tags or with db:"-" are ignored.
func InsertStruct(ctx context.Context, db DB, table string, data any) error {
	fields, values, err := extractStructFields(data)
	if err != nil {
		return fmt.Errorf("failed to extract struct fields: %w", err)
	}

	if len(fields) == 0 {
		return fmt.Errorf("no valid fields found for insertion")
	}

	placeholders := make([]string, len(fields))
	for i := range fields {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err = db.Exec(ctx, sql, values...)
	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	return nil
}

// QueryStructs executes a query and maps results into the provided struct slice.
// It uses db:"table.column" tags to map columns to struct fields.
// The dest parameter must be a pointer to a slice of structs.
func QueryStructs(ctx context.Context, db DB, sql string, dest any, args ...any) error {
	fmt.Printf("[dbx] QueryStructs called with dest type: %T, value: %#v\n", dest, dest)
	destValue := reflect.ValueOf(dest)
	if dest == nil {
		return fmt.Errorf("dest cannot be nil; must be a pointer to a slice of structs")
	}
	if destValue.Kind() != reflect.Pointer {
		return fmt.Errorf("dest must be a pointer to a slice of structs, got %T", dest)
	}
	if destValue.IsNil() {
		return fmt.Errorf("dest pointer is nil; must be a pointer to a slice of structs")
	}

	sliceValue := destValue.Elem()
	if sliceValue.Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to a slice of structs, got pointer to %s", sliceValue.Kind())
	}

	// Get the element type of the slice
	elemType := sliceValue.Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("slice elements must be structs, got %s", elemType.Kind())
	}

	// Execute the query
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Build field mapping
	fieldMap, err := buildFieldMapping(rows, elemType)
	if err != nil {
		return fmt.Errorf("failed to build field mapping: %w", err)
	}

	// Process each row
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return fmt.Errorf("failed to get row values: %w", err)
		}

		// Create a new struct instance
		elem := reflect.New(elemType).Elem()

		// Map values to struct fields
		for colIndex, fieldIndex := range fieldMap {
			if colIndex < len(values) && fieldIndex >= 0 {
				field := elem.Field(fieldIndex)
				if field.CanSet() {
					val := reflect.ValueOf(values[colIndex])
					if !val.IsValid() || (val.Kind() == reflect.Ptr && val.IsNil()) {
						// Set zero value for the field if DB value is NULL
						field.Set(reflect.Zero(field.Type()))
					} else if val.Type().ConvertibleTo(field.Type()) {
						field.Set(val.Convert(field.Type()))
					}
				}
			}
		}

		// Append to the slice
		sliceValue.Set(reflect.Append(sliceValue, elem))
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("row iteration error: %w", err)
	}

	return nil
}

// extractStructFields extracts field names and values from a struct for insertion.
// It uses db tags to determine column names and skips fields with db:"-".
func extractStructFields(data any) ([]string, []any, error) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("data must be a struct or pointer to struct")
	}

	t := v.Type()
	var fields []string
	var values []any

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")

		// Skip fields with no db tag or explicitly ignored
		if dbTag == "" || dbTag == "-" {
			continue
		}

		// Extract the column name from the tag
		// Support both "column" and "table.column" formats
		columnName := dbTag
		if dotIndex := strings.Index(dbTag, "."); dotIndex != -1 {
			columnName = dbTag[dotIndex+1:]
		}

		fields = append(fields, columnName)
		values = append(values, v.Field(i).Interface())
	}

	return fields, values, nil
}

// buildFieldMapping creates a mapping from column indices to struct field indices.
// It uses db tags to match columns to fields, with fallback to field names.
func buildFieldMapping(rows pgx.Rows, structType reflect.Type) (map[int]int, error) {
	fieldDescs := rows.FieldDescriptions()
	fieldMap := make(map[int]int)

	// Build a map of column names to their indices
	colMap := make(map[string]int)
	for i, fd := range fieldDescs {
		colMap[string(fd.Name)] = i
	}

	// Map struct fields to columns
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		dbTag := field.Tag.Get("db")

		if dbTag == "" || dbTag == "-" {
			continue
		}

		// Try to find the column by the full tag first
		if colIndex, exists := colMap[dbTag]; exists {
			fieldMap[colIndex] = i
			continue
		}

		// If it's a table.column format, try just the column name
		if dotIndex := strings.Index(dbTag, "."); dotIndex != -1 {
			columnName := dbTag[dotIndex+1:]
			if colIndex, exists := colMap[columnName]; exists {
				fieldMap[colIndex] = i
				continue
			}
		}

		// Fallback to field name
		if colIndex, exists := colMap[field.Name]; exists {
			fieldMap[colIndex] = i
		}
	}

	return fieldMap, nil
}
