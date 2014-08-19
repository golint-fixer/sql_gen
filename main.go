package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	tableRegex     = regexp.MustCompile(`CREATE TABLE (\w+) \(((?s).*?)\);((?s).*\z)`)
	goFileRegex    = regexp.MustCompile(`(.+).go`)
	goPackageRegex = regexp.MustCompile(`package (\w+)`)
)

type goType struct {
	name         string // the name of the golang type to be used
	fn           string // the string statement to be used in the insert statement
	importNeeded string // for necessary golang imports
}

// Schema is used to keep track of all information about a sql database schema.
type Schema struct {
	Name    string
	Imports map[string]interface{}
	Columns []Column
}

// Column represents a sql column in it's corresponding Golang struct form
type Column struct {
	Attr     string
	Name     string
	DataType goType
}

func getSchemaData(schema string) (Schema, string, error) {
	result := tableRegex.FindStringSubmatch(schema)
	if len(result) != 4 {
		return Schema{}, "", errors.New("Failed to find schema in file provided")
	}
	imports, columns := parseColumns(result[2])
	return Schema{
		Name:    result[1],
		Imports: imports,
		Columns: columns,
	}, result[3], nil
}

// creates an array of Columns given a PG string of columns
func parseColumns(columnStr string) (map[string]interface{}, []Column) {
	columns := strings.Split(columnStr, `,`)
	tableColumns := make([]Column, len(columns)) // we know the exact size
	imports := make(map[string]interface{})

	for i, c := range columns {
		data := strings.SplitN(strings.Trim(c, " \n"), " ", 2)
		name, pgType := data[0], data[1]
		golangType := translatePGType(pgType)
		if golangType.name == "" {
			log.Fatalf("DataType %s not yet supported\n", pgType)
		}

		tableColumns[i] = Column{
			// make sure the golang attribute starts with an uppercase letter so it's public
			Attr:     strings.ToUpper(name[0:1]) + name[1:],
			Name:     name,
			DataType: golangType,
		}
		if golangType.importNeeded != "" {
			imports[golangType.importNeeded] = 0
		}
	}
	return imports, tableColumns
}

func translatePGType(pgType string) goType {
	switch pgType {
	case "text",
		"character varying(32)",
		"character varying(64)",
		"character varying(32) NOT NULL",
		"character varying(64) NOT NULL":
		return goType{
			name: "string",
			fn:   "%s"}
	case "boolean":
		return goType{
			name: "bool",
			fn:   "%s"}
	case "double precision":
		return goType{
			name: "float64",
			fn:   "%s"}
	case "time without time zone":
		return goType{
			name:         "time.Time",
			fn:           `%s.Format("15:04")`,
			importNeeded: `"time"`,
		}
	case "integer":
		return goType{
			name: "int",
			fn:   "%s",
		}
	}
	return goType{}
}

// finds the name of the go package for this directory
func getGoPackage() (string, error) {
	// examine all .go files
	files, err := filepath.Glob("*.go")
	if err != nil {
		return "", fmt.Errorf("Failed to read files in directory => %s", err.Error())
	}
	var returnErr error

	// iterate over all matches, will process files until a match is found
	for _, f := range files {
		// read in the file
		fileBytes, err := ioutil.ReadFile(f)
		if err != nil {
			returnErr = fmt.Errorf("IO error reading in go file, %s", err.Error())
			continue
		}

		// go fmt the file for consistency
		srcBytes, err := format.Source(fileBytes)
		if err != nil {
			returnErr = fmt.Errorf("Go FMT error formating go file, %s", err.Error())
			continue
		}

		// read in the first line of the go file
		var buf = bytes.NewBuffer(srcBytes)
		packageLine, err := buf.ReadString('\n')
		if err != nil {
			returnErr = fmt.Errorf("IO error while reading first line of go file, %s", err.Error())
			continue
		}

		// find the name of the package via regex
		matches := goPackageRegex.FindStringSubmatch(packageLine)
		if len(matches) != 2 {
			returnErr = fmt.Errorf("Error while parsing first line of go file, could not find package")
			continue
		}
		return matches[1], nil
	}

	// default to main
	return "main", returnErr
}

func readInSchema(src io.Reader) []Schema {
	stdinBytes, err := ioutil.ReadAll(src)
	if err != nil {
		log.Fatalf("Failed to read in stdin => %s", err.Error())
	}

	var s Schema
	schemaString := string(stdinBytes)
	schemas := []Schema{}
	for {
		s, schemaString, err = getSchemaData(schemaString)
		if err != nil {
			break
		}
		schemas = append(schemas, s)
	}

	return schemas
}

func main() {
	s := readInSchema(os.Stdin)
	if len(s) == 0 {
		log.Fatal("No sql tables schemas found.")
	}

	for _, schema := range s {
		log.Printf("Generating functions for %s", schema.Name)
		err := generateGoFile(schema, fmt.Sprintf("./%s_sql.go", schema.Name))
		if err != nil {
			log.Fatalf("Failed to generate file => %s", err.Error())
		}
	}
}
