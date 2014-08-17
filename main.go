package main

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"regexp"
	"strings"
)

var (
	tableRegex = regexp.MustCompile(`CREATE TABLE (\w+) \(((?s).*)\);`)
	pgConv     = map[string]string{
		"text":                   "string",
		"character varying(32)":  "string",
		"character varying(64)":  "string",
		"time without time zone": "string",
		"integer":                "int",
	}
)

// Schema is used to keep track of all information about a sql database schema.
type Schema struct {
	Name    string
	Columns []Column
}

// Column represents a sql column in it's corresponding Golang struct form
type Column struct {
	Attr, Name, DataType string
}

func generateStruct(s Schema) (string, error) {
	attributes := ""
	for _, c := range s.Columns {
		attributes += fmt.Sprintf("%s %s\n", c.Attr, c.DataType)
	}
	return gofmt(`type %s struct {%s}`, s.Name, attributes)
}

func generateInsert(s Schema) (string, error) {
	// form variable parts of statement
	abbrev := s.Name[0:1]
	attributes := s.Columns[0].Name
	sqlParameters := "$1"
	parameters := fmt.Sprintf("%s.%s", abbrev, s.Columns[0].Attr)
	for i, c := range s.Columns[1:] {
		attributes += fmt.Sprintf(", %s", c.Name)
		// offset 1 for sql, another 1 for starting on the 2nd elem
		sqlParameters += fmt.Sprintf(", $%d", i+2)
		parameters += fmt.Sprintf(", %s.%s", abbrev, c.Attr)
	}

	return gofmt(`
func (%s %s) Insert (db *sql.DB) error {
	query := "INSERT INTO %s (%s) VALUES (%s)"
	_, err := db.Exec(query, %s)
	if err != nil {
		return fmt.Errorf("Failed to insert Course, %%#v, => %%s", c, err.Error())
	}
	return nil
}`, abbrev, s.Name, s.Name, attributes, sqlParameters, parameters)
}

func gofmt(src string, args ...interface{}) (string, error) {
	var buf bytes.Buffer

	n, err := fmt.Fprintf(&buf, src, args...)
	if err != nil || n == 0 {
		return "", fmt.Errorf("Failed to properly render code while generating")
	}

	rendered, err := format.Source(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("Failed to properly fmt code while generating")
	}
	return string(rendered), nil
}

// creates an array of Columns given a PG string of columns
func parseColumns(columnStr string) []Column {
	columns := strings.Split(columnStr, `,`)
	tableColumns := make([]Column, len(columns)) // we know the exact size

	for i, c := range columns {
		data := strings.SplitN(strings.Trim(c, " \n"), " ", 2)
		name, pgType := data[0], data[1]
		goType, match := pgConv[pgType]
		if !match {
			log.Fatalf("DataType %s not yet supported\n", pgType)
		}
		tableColumns[i] = Column{
			// make sure the golang attribute starts with an uppercase letter
			Attr:     strings.ToUpper(name[0:1]) + name[1:],
			Name:     name,
			DataType: goType,
		}
	}
	return tableColumns
}

func getSchemaData(schema string) (Schema, error) {
	result := tableRegex.FindStringSubmatch(schema)
	if len(result) != 3 {
		log.Fatal("Failed to find schema in file provided")
	}
	return Schema{
		Name:    result[1],
		Columns: parseColumns(result[2]),
	}, nil
}

func main() {
}
