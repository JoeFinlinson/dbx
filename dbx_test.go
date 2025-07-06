package dbx

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Mock implementation for testing
type mockQueryer struct {
	rows []mockRow
}

type mockRow struct {
	values []interface{}
}

func (m *mockQueryer) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return &mockRows{rows: m.rows, current: -1}, nil
}

func (m *mockQueryer) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

type mockRows struct {
	rows    []mockRow
	current int
}

func (m *mockRows) Next() bool {
	m.current++
	return m.current < len(m.rows)
}

func (m *mockRows) Scan(dest ...interface{}) error {
	return nil
}

func (m *mockRows) Values() ([]interface{}, error) {
	if m.current >= 0 && m.current < len(m.rows) {
		return m.rows[m.current].values, nil
	}
	return nil, nil
}

func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription {
	return []pgconn.FieldDescription{
		{Name: "id"},
		{Name: "name"},
		{Name: "email"},
	}
}

func (m *mockRows) Close() {}

func (m *mockRows) Err() error {
	return nil
}

func (m *mockRows) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}

func (m *mockRows) RawValues() [][]byte {
	return nil
}

func (m *mockRows) Conn() *pgx.Conn {
	return nil
}

func TestQueryMaps(t *testing.T) {
	ctx := context.Background()
	mock := &mockQueryer{
		rows: []mockRow{
			{values: []interface{}{1, "John", "john@example.com"}},
			{values: []interface{}{2, "Jane", "jane@example.com"}},
		},
	}

	results, err := QueryMaps(ctx, mock, "SELECT * FROM users")
	if err != nil {
		t.Fatalf("QueryMaps failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(results))
	}

	expected := RowMap{"id": 1, "name": "John", "email": "john@example.com"}
	if !reflect.DeepEqual(results[0], expected) {
		t.Errorf("Expected %+v, got %+v", expected, results[0])
	}
}

func TestQueryJSON(t *testing.T) {
	ctx := context.Background()
	mock := &mockQueryer{
		rows: []mockRow{
			{values: []interface{}{1, "John", "john@example.com"}},
		},
	}

	jsonData, err := QueryJSON(ctx, mock, "SELECT * FROM users")
	if err != nil {
		t.Fatalf("QueryJSON failed: %v", err)
	}

	var result []RowMap
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(result))
	}
}

func TestInsertStruct(t *testing.T) {
	ctx := context.Background()
	mock := &mockQueryer{}

	type TestUser struct {
		Name  string `db:"name"`
		Email string `db:"email"`
		Age   int    `db:"-"` // Should be ignored
	}

	user := TestUser{
		Name:  "Test User",
		Email: "test@example.com",
		Age:   30, // Should be ignored
	}

	err := InsertStruct(ctx, mock, "users", user)
	if err != nil {
		t.Fatalf("InsertStruct failed: %v", err)
	}
}

func TestQueryStructs(t *testing.T) {
	ctx := context.Background()
	mock := &mockQueryer{
		rows: []mockRow{
			{values: []interface{}{1, "John", "john@example.com"}},
			{values: []interface{}{2, "Jane", "jane@example.com"}},
		},
	}

	type TestUser struct {
		ID    int    `db:"id"`
		Name  string `db:"name"`
		Email string `db:"email"`
	}

	var users []TestUser
	err := QueryStructs(ctx, mock, "SELECT * FROM users", &users)
	if err != nil {
		t.Fatalf("QueryStructs failed: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("Expected 2 users, got %d", len(users))
	}

	if users[0].Name != "John" || users[0].Email != "john@example.com" {
		t.Errorf("First user data incorrect: %+v", users[0])
	}
}

func TestQueryStructsWithTableColumnTags(t *testing.T) {
	ctx := context.Background()
	mock := &mockQueryer{
		rows: []mockRow{
			{values: []interface{}{1, "John", "john@example.com"}},
		},
	}

	type TestUser struct {
		Users_ID    int    `db:"users.id"`
		Users_Name  string `db:"users.name"`
		Users_Email string `db:"users.email"`
	}

	var users []TestUser
	err := QueryStructs(ctx, mock, "SELECT * FROM users", &users)
	if err != nil {
		t.Fatalf("QueryStructs with table.column tags failed: %v", err)
	}

	if len(users) != 1 {
		t.Fatalf("Expected 1 user, got %d", len(users))
	}
}

func TestExtractStructFields(t *testing.T) {
	type TestStruct struct {
		Name  string `db:"name"`
		Email string `db:"email"`
		Age   int    `db:"-"`
		Title string // No db tag
	}

	data := TestStruct{
		Name:  "John",
		Email: "john@example.com",
		Age:   30,
		Title: "Mr",
	}

	fields, values, err := extractStructFields(data)
	if err != nil {
		t.Fatalf("extractStructFields failed: %v", err)
	}

	expectedFields := []string{"name", "email"}
	if !reflect.DeepEqual(fields, expectedFields) {
		t.Errorf("Expected fields %v, got %v", expectedFields, fields)
	}

	expectedValues := []interface{}{"John", "john@example.com"}
	if !reflect.DeepEqual(values, expectedValues) {
		t.Errorf("Expected values %v, got %v", expectedValues, values)
	}
}

func TestExtractStructFieldsWithTableColumn(t *testing.T) {
	type TestStruct struct {
		Name  string `db:"users.name"`
		Email string `db:"users.email"`
	}

	data := TestStruct{
		Name:  "John",
		Email: "john@example.com",
	}

	fields, _, err := extractStructFields(data)
	if err != nil {
		t.Fatalf("extractStructFields failed: %v", err)
	}

	expectedFields := []string{"name", "email"}
	if !reflect.DeepEqual(fields, expectedFields) {
		t.Errorf("Expected fields %v, got %v", expectedFields, fields)
	}
}

func TestBuildFieldMapping(t *testing.T) {
	mock := &mockQueryer{
		rows: []mockRow{
			{values: []interface{}{1, "John", "john@example.com"}},
		},
	}

	rows, _ := mock.Query(context.Background(), "SELECT * FROM users")
	defer rows.Close()

	type TestUser struct {
		ID    int    `db:"id"`
		Name  string `db:"name"`
		Email string `db:"email"`
	}

	fieldMap, err := buildFieldMapping(rows, reflect.TypeOf(TestUser{}))
	if err != nil {
		t.Fatalf("buildFieldMapping failed: %v", err)
	}

	// Should map column 0 (id) to field 0, column 1 (name) to field 1, etc.
	expectedMap := map[int]int{0: 0, 1: 1, 2: 2}
	if !reflect.DeepEqual(fieldMap, expectedMap) {
		t.Errorf("Expected field map %v, got %v", expectedMap, fieldMap)
	}
}

func TestBuildFieldMappingWithTableColumn(t *testing.T) {
	mock := &mockQueryer{
		rows: []mockRow{
			{values: []interface{}{1, "John", "john@example.com"}},
		},
	}

	rows, _ := mock.Query(context.Background(), "SELECT * FROM users")
	defer rows.Close()

	type TestUser struct {
		Users_ID    int    `db:"users.id"`
		Users_Name  string `db:"users.name"`
		Users_Email string `db:"users.email"`
	}

	fieldMap, err := buildFieldMapping(rows, reflect.TypeOf(TestUser{}))
	if err != nil {
		t.Fatalf("buildFieldMapping with table.column failed: %v", err)
	}

	// Should map columns to fields using the column name part after the dot
	expectedMap := map[int]int{0: 0, 1: 1, 2: 2}
	if !reflect.DeepEqual(fieldMap, expectedMap) {
		t.Errorf("Expected field map %v, got %v", expectedMap, fieldMap)
	}
}
