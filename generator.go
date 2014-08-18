package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
)

func generateStruct(s Schema) (string, error) {
	attributes := ""
	for _, c := range s.Columns {
		attributes += fmt.Sprintf("%s %s\n", c.Attr, c.DataType.name)
	}
	return gofmt(`type %s struct {%s}`, s.Name, attributes)
}

// generates a method for inserting a struct into a provided DB
func generateInsert(s Schema) (string, error) {
	// form variable parts of statement
	abbrev := s.Name[0:1]
	attributes := s.Columns[0].Name
	sqlParameters := "$1"
	parameters := fmt.Sprintf(
		"%s.%s,\n",
		abbrev,
		fmt.Sprintf(s.Columns[0].DataType.fn, s.Columns[0].Attr),
	)
	for i, c := range s.Columns[1:] {
		attributes += fmt.Sprintf(", %s", c.Name)
		// offset 1 for sql, another 1 for starting on the 2nd elem
		sqlParameters += fmt.Sprintf(", $%d", i+2)
		parameters += fmt.Sprintf(c.DataType.fn, fmt.Sprintf("%s.%s", abbrev, c.Attr)) + ",\n"
	}

	return gofmt(`
func (%s %s) Insert (db *sql.DB) error {
	query := "INSERT INTO %s (%s) VALUES (%s)"
	_, err := db.Exec(
		query,
		%s
	)
	if err != nil {
		return fmt.Errorf("Failed to insert %s, %%#v, => %%s", %s, err.Error())
	}
	return nil
}`, abbrev, s.Name, s.Name, attributes, sqlParameters, parameters, s.Name, abbrev)
}

// generates a method for scanning a sql.Row into the generated struct
func generateScan(s Schema) (string, error) {
	// form variable parts of statement
	abbrev := s.Name[0:1]
	parameters := fmt.Sprintf("&%s.%s,", abbrev, s.Columns[0].Attr)
	for _, c := range s.Columns[1:] {
		parameters += fmt.Sprintf("\n&%s.%s,", abbrev, c.Attr)
	}
	return gofmt(`
func (%s *%s) Scan(row *sql.Row) error {
	return row.Scan(
		%s
	)
}`, abbrev, s.Name, parameters)
}

func generateImports(s Schema) (string, error) {
	imports := ""
	for i := range s.Imports {
		imports += i + "\n"
	}

	return fmt.Sprintf(`
import (
	"database/sql"
	"fmt"
	%s
)`, imports), nil
}

func generateFileString(s Schema) (string, error) {
	// get the name of the golang package
	packageName, err := getGoPackage()
	if err != nil {
		return "", fmt.Errorf(`Failed to find package name, defaulting to "main"`)
	}

	// get the imports
	imports, err := generateImports(s)
	if err != nil {
		return "", fmt.Errorf("Error while generating imports => %s", err.Error())
	}

	// get struct
	structString, err := generateStruct(s)
	if err != nil {
		return "", fmt.Errorf("Error while generating struct => %s", err.Error())
	}

	// get sql row scanner
	scanString, err := generateScan(s)
	if err != nil {
		return "", fmt.Errorf("Error while generating Scan method => %s", err.Error())
	}

	// get sql insert
	insertString, err := generateInsert(s)
	if err != nil {
		return "", fmt.Errorf("Error while generating Insert method => %s", err.Error())
	}

	return gofmt(`
package %s
%s
%s
%s
%s`, packageName, imports, structString, scanString, insertString)
}

func generateGoFile(s Schema, path string) error {
	fileString, err := generateFileString(s)
	if err != nil {
		return fmt.Errorf("Failed to generate go file string => %s", err.Error())
	}

	buf := bytes.NewBufferString(fileString)
	err = ioutil.WriteFile(path, buf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("Failed to write go file => %s", err.Error())
	}
	return nil
}

// string formats with the arguments then adds itself to gofmt standard
func gofmt(src string, args ...interface{}) (string, error) {
	var buf bytes.Buffer

	n, err := fmt.Fprintf(&buf, src, args...)
	if err != nil || n == 0 {
		return "", fmt.Errorf("Failed to properly render code while generating, %s", err.Error())
	}

	// fmt.Println(buf.String())
	rendered, err := format.Source(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("Failed to properly fmt code while generating, %s", err.Error())
	}
	return string(rendered), nil
}
