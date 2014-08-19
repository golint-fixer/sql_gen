package main

import (
	"bytes"
	"testing"
)

func TestGetData(t *testing.T) {
	schema, _, err := getSchemaData(testSchema)
	if err != nil {
		t.Fatal(err)
	}

	if schema.Name != expectedName {
		t.Errorf("incorrectly found table name, found %s", schema.Name)
	}
	if schema.Columns[0] != expectedColumn0 {
		t.Errorf("improperly parsed column 0, should be %#v", expectedColumn0)
		t.Errorf("Recieved: %#v\n", schema.Columns[0])
	}
}

func TestParseColumns(t *testing.T) {
	_, columns := parseColumns(testColumnString)
	if columns[0] != expectedColumn0 {
		t.Errorf("improperly parsed column 0, should be %#v", expectedColumn0)
		t.Errorf("Recieved: %#v\n", columns[0])
	}
}

func TestGenStruct(t *testing.T) {
	schema, _, err := getSchemaData(testSchema)
	if err != nil {
		t.Fatal(err)
	}

	structString, err := generateStruct(schema)
	if err != nil {
		t.Fatal(err)
	}

	if structString != expectedStruct {
		t.Errorf("Improperly generated struct string")
		t.Errorf("EXPECTED: %s\n", expectedStruct)
		t.Errorf("RECEIVED: %s\n", structString)
	}
}

func TestGenInsert(t *testing.T) {
	schema, _, err := getSchemaData(testSchema)
	if err != nil {
		t.Fatal(err)
	}

	insertFn, err := generateInsert(schema)
	if err != nil {
		t.Errorf(insertFn)
		t.Fatal(err)
	}

	if insertFn != expectedInsertMethod {
		t.Errorf("Improperly generated insert method string")
		t.Errorf("EXPECTED: %s\n", expectedInsertMethod)
		t.Errorf("RECEIVED: %s\n", insertFn)
	}
}

func TestGenScan(t *testing.T) {
	schema, _, _ := getSchemaData(testSchema)
	scanFn, err := generateScan(schema)
	if err != nil {
		t.Error(err)
	}
	if scanFn != expectedScanMethod {
		t.Errorf("Improperly generated scan method string")
		t.Errorf("EXPECTED: %s\n", expectedScanMethod)
		t.Errorf("RECEIVED: %s\n", scanFn)
	}
}

func TestGetGoPackage(t *testing.T) {
	name, _ := getGoPackage()
	if name != expectedPackage {
		t.Errorf("Improperly found go package")
		t.Errorf("EXPECTED: %s\n", expectedPackage)
		t.Errorf("RECEIVED: %s\n", name)
	}
}

func TestMultipleSchemas(t *testing.T) {
	schemas := readInSchema(bytes.NewBufferString(testSchema + testSchema2))
	if len(schemas) != 2 {
		t.Errorf("Should've parsed 2 schemas")
	}
}

var (
	testSchema = `CREATE TABLE courses_t (
    term character varying(32),
    callnumber integer,
    classnotes character varying(64),
    starttime1 time without time zone,
    description text
);`
	testSchema2 = `CREATE TABLE users_t (
	email character varying(64) NOT NULL,
	token character varying(32) NOT NULL,
	name character varying(64) NOT NULL
);`
	testColumnString = `
    term character varying(32),
    callnumber integer,
    classnotes character varying(64),
    starttime1 time without time zone,
    description text
`
	expectedStruct = `type courses_t struct {
	Term        string
	Callnumber  int
	Classnotes  string
	Starttime1  time.Time
	Description string
}`
	expectedInsertMethod = `
func (c courses_t) Insert(db *sql.DB) error {
	query := "INSERT INTO courses_t (term, callnumber, classnotes, starttime1, description) VALUES ($1, $2, $3, $4, $5)"
	_, err := db.Exec(
		query,
		c.Term,
		c.Callnumber,
		c.Classnotes,
		c.Starttime1.Format("15:04"),
		c.Description,
	)
	if err != nil {
		return fmt.Errorf("Failed to insert courses_t, %#v, => %s", c, err.Error())
	}
	return nil
}`
	expectedScanMethod = `
func (c *courses_t) Scan(row *sql.Row) error {
	return row.Scan(
		&c.Term,
		&c.Callnumber,
		&c.Classnotes,
		&c.Starttime1,
		&c.Description,
	)
}`
	expectedName    = "courses_t"
	expectedPackage = "main"
	expectedColumn0 = Column{
		Attr:     "Term",
		Name:     "term",
		DataType: goType{"string", "%s", ""},
	}
)
