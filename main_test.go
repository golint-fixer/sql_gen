package main

import (
	"strings"
	"testing"
)

func TestGetData(t *testing.T) {
	tableName, columns := getSchemaData(testSchema)
	if tableName != expectedName {
		t.Errorf("incorrectly found table name, found %s", tableName)
	}
	if !strings.Contains(columns, `bulletinflags character varying(32),`) {
		t.Errorf("improperly found columns, should contain all columns")
	}
}

func TestParseColumns(t *testing.T) {
	_, col := getSchemaData(testSchema)
	columns := parseColumns(col)
	if columns[0] != expectedColumn0 {
		t.Errorf("improperly parsed column 0, should be %#v", expectedColumn0)
	}
}

func TestGenStruct(t *testing.T) {
	name, cols := getSchemaData(testSchema)
	columns := parseColumns(cols)
	structString := generateStruct(name, columns)
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
	expectedStruct = `type courses_t struct {
	term          string
	callnumber    int
	bulletinflags string
	classnotes    string
	starttime1    string
	description   string
}`
	expectedName    = "courses_t"
	expectedColumn0 = Column{
		Name:     "term",
		DataType: "string",
	}
)
