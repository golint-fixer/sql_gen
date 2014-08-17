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
	structString := generateStruct(schema)
	if structString != expectedStruct {
		t.Errorf("Improperly generated struct string")
	}
}

var (
	testSchema = `CREATE TABLE courses_t (
    term character varying(32),
    callnumber integer,
    bulletinflags character varying(32),
    classnotes character varying(64),
    starttime1 time without time zone,
    description text
);`
	testColumnString = `
    term character varying(32),
    callnumber integer,
    bulletinflags character varying(32),
    classnotes character varying(64),
    starttime1 time without time zone,
    description text
`
	expectedStruct = `type courses_t struct {
	Term          string
	Callnumber    int
	Bulletinflags string
	Classnotes    string
	Starttime1    string
	Description   string
}`
	expectedName    = "courses_t"
	expectedColumn0 = Column{
		Attr:     "Term",
		Name:     "term",
		DataType: "string",
	}
)
