package main

import "testing"

func TestGetData(t *testing.T) {
	schema, _ := getSchemaData(testSchema)
	if schema.Name != expectedName {
		t.Errorf("incorrectly found table name, found %s", schema.Name)
	}
	if schema.Columns[0] != expectedColumn0 {
		t.Errorf("improperly parsed column 0, should be %#v", expectedColumn0)
		t.Errorf("Recieved: %#v\n", schema.Columns[0])
	}
}

func TestParseColumns(t *testing.T) {
	columns := parseColumns(testColumnString)
	if columns[0] != expectedColumn0 {
		t.Errorf("improperly parsed column 0, should be %#v", expectedColumn0)
		t.Errorf("Recieved: %#v\n", columns[0])
	}
}

func TestGenStruct(t *testing.T) {
	schema, _ := getSchemaData(testSchema)
	structString, _ := generateStruct(schema)
	if structString != expectedStruct {
		t.Errorf("Improperly generated struct string")
		t.Errorf("EXPECTED: %s\n", expectedStruct)
		t.Errorf("RECEIVED: %s\n", structString)
	}
}

func TestGenInsert(t *testing.T) {
	schema, _ := getSchemaData(testSchema)
	insertFn, _ := generateInsert(schema)
	if insertFn != expectedInsertMethod {
		t.Errorf("Improperly generated insert method string")
		t.Errorf("EXPECTED: %s\n", expectedInsertMethod)
		t.Errorf("RECEIVED: %s\n", insertFn)
	}
}

func TestGenScan(t *testing.T) {
	schema, _ := getSchemaData(testSchema)
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

var (
	testSchema = `CREATE TABLE courses_t (
    term character varying(32),
    callnumber integer,
    classnotes character varying(64),
    description text
);`
	testColumnString = `
    term character varying(32),
    callnumber integer,
    classnotes character varying(64),
    description text
`
	expectedStruct = `type courses_t struct {
	Term        string
	Callnumber  int
	Classnotes  string
	Description string
}`
	expectedInsertMethod = `
func (c courses_t) Insert(db *sql.DB) error {
	query := "INSERT INTO courses_t (term, callnumber, classnotes, description) VALUES ($1, $2, $3, $4)"
	_, err := db.Exec(query, c.Term, c.Callnumber, c.Classnotes, c.Description)
	if err != nil {
		return fmt.Errorf("Failed to insert Course, %#v, => %s", c, err.Error())
	}
	return nil
}`
	expectedScanMethod = `
func (c courses_t) Scan(row *sql.Row) error {
	return row.Scan(&c.Term, &c.Callnumber, &c.Classnotes, &c.Description)
}`
	expectedName    = "courses_t"
	expectedColumn0 = Column{
		Attr:     "Term",
		Name:     "term",
		DataType: "string",
	}
)
