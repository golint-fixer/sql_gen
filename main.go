package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	tableRegex     = regexp.MustCompile(`CREATE TABLE (\w+) \(((?s).*)\);`)
	goFileRegex    = regexp.MustCompile(`(.+).go`)
	goPackageRegex = regexp.MustCompile(`package (\w+)`)
	pgConv         = map[string]string{
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
			// make sure the golang attribute starts with an uppercase letter so it's public
			Attr:     strings.ToUpper(name[0:1]) + name[1:],
			Name:     name,
			DataType: goType,
		}
	}
	return tableColumns
}

// finds the name of the go package for this directory
func getGoPackage() (string, error) {
	// examine all .go files
	files, err := filepath.Glob("*.go")
	if err != nil {
		return "", fmt.Errorf("Failed to read files in directory => %s", err.Error())
	}
	var returnErr error = nil

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

func readStdin() Schema {
	stdinBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("Failed to read in stdin => %s", err.Error())
	}

	s, err := getSchemaData(string(stdinBytes))
	if err != nil {
		log.Fatalf("Failed to parse stdin into schema => %s", err.Error())
	}

	return s
}

func main() {
	s := readStdin()
	err := generateGoFile(s, fmt.Sprintf("./%s_sql.go", s.Name))
	if err != nil {
		log.Fatalf("Failed to generate file => %s", err.Error())
	}
}
