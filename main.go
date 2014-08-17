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

type Schema struct {
	Name    string
	Columns []Column
}

// Column represents a sql column in it's corresponding Golang struct form
type Column struct {
	Attr, Name, DataType string
}

func generateStruct(s Schema) string {
	attributes := ""
	for _, c := range s.Columns {
		attributes += fmt.Sprintf("%s %s\n", c.Attr, c.DataType)
	}
	var buf bytes.Buffer
	n, err := fmt.Fprintf(&buf, `type %s struct {%s}`, s.Name, attributes)
	if err != nil || n == 0 {
		log.Fatal("Failed to properly render code while generating struct")
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal("Failed to properly fmt code while generating struct")
	}
	return string(src)
}

func generateInsert(s Schema) string {
	// form variable parts of statement
	abbrev := s.Name[0]
	attributes := s.Columns[0].Name
	sqlParameters := "$1"
	parameters := fmt.Sprintf("%s.%s", abbrev, s.Columns[0].Name)
	for i, c := range s.Columns[1:] {
		attributes += fmt.Sprintf(", %s", c.Name)
		sqlParameters += fmt.Sprintf(", $%d", i)
		parameters += fmt.Sprintf(", %s.%s", abbrev, c.Attr)
	}

	var buf bytes.Buffer
	n, err := fmt.Fprintf(&buf, `
func (%s %s) Insert (db *db.Sql) {
	query := "INSERT INTO %s (%s) VALUES %s"
	_, err := db.Exec(query, %s)
	if err != nil {
		return fmt.Errorf("Failed to insert Course, %%#v, => %%s", c, err.Error())
	}
	return nil
}`, abbrev, s.Name, s.Name, attributes, sqlParameters, parameters)
	if err != nil || n == 0 {
		log.Fatal("Failed to properly render code while generating insert statement")
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal("Failed to properly fmt code while generating insert statement")
	}
	return string(src)
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
