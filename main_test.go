package main

import (
	"strings"
	"testing"
)

func TestGetTableName(t *testing.T) {
	tableName := getTableName(testSchema)
	if tableName != expectedName {
		t.Errorf("incorrectly found table name, found %s", tableName)
	}
}

func TestGetColumns(t *testing.T) {
	columns := getColumns(testSchema)
	if !strings.Contains(columns, `bulletinflags character varying(32),`) {
		t.Errorf("improperly found columns, should contain all columns")
	}
}

func TestParseColumns(t *testing.T) {
	columns := parseColumns(getColumns(testSchema))
	if columns[0] != expectedColumn0 {
		t.Errorf("improperly parsed column 0, should be %#v", expectedColumn0)
	}
}

var (
	testSchema = `CREATE TABLE courses_t (
    term character varying(32),
    callnumber integer,
    bulletinflags character varying(32),
    classnotes character varying(64),
    starttime1 time without time zone,
    endtime1 time without time zone,
    description text
);`
	expectedName    = "courses_t"
	expectedColumn0 = Column{
		name:     "term",
		datatype: "string",
	}
)
