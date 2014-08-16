package main

import (
	"fmt"
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
	name, datatype string
}

func parseColumns(columnStr string) []Column {
	columns := strings.Split(columnStr, `,`)
	tableColumns := make([]Column, len(columns))
	for i, c := range columns {
		data := strings.SplitN(strings.Trim(c, " \n"), " ", 2)
		name, pgType := data[0], data[1]
		goType, match := pgConv[pgType]
		if !match {
			panic(fmt.Sprintf("Datatype %s not yet supported\n", pgType))
		}
		tableColumns[i] = Column{
			name:     name,
			datatype: goType,
		}
	}
	return tableColumns
}

func getColumns(schema string) string {
	result := tableRegex.FindStringSubmatch(schema)
	if len(result) != 3 {
		return ""
	}
	return result[2]
}

func getTableName(schema string) string {
	result := tableRegex.FindStringSubmatch(schema)
	if len(result) != 3 {
		return ""
	}
	return result[1]
}

func main() {

}
