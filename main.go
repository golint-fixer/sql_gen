package main

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"regexp"
	"strings"
)

const tableRegexStr = `CREATE TABLE (\w+) \(((?s).*)\);`

var (
	tableRegex = regexp.MustCompile(tableRegexStr)
	pgConv     = map[string]string{
		"text":                   "string",
		"character varying(32)":  "string",
		"character varying(64)":  "string",
		"time without time zone": "string",
		"integer":                "int",
	}
)

// Column represents a sql column in it's corresponding Golang struct form
type Column struct {
	Name, DataType string
}

func generateStruct(tableName string, columns []Column) string {
	attributes := ""
	for _, c := range columns {
		attributes += fmt.Sprintf("%s %s\n", c.Name, c.DataType)
	}
	var buf bytes.Buffer
	n, err := fmt.Fprintf(&buf, `type %s struct {%s}`, tableName, attributes)
	if err != nil || n == 0 {
		log.Fatal("Failed to properly render code while generating struct")
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal("Failed to properly fmt code while generating struct")
		return ""
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
			panic(fmt.Sprintf("DataType %s not yet supported\n", pgType))
		}
		tableColumns[i] = Column{
			Name:     name,
			DataType: goType,
		}
	}
	return tableColumns
}

func getSchemaData(schema string) (string, string) {
	result := tableRegex.FindStringSubmatch(schema)
	if len(result) != 3 {
		return "", ""
	}
	return result[1], result[2]
}

func main() {
}
