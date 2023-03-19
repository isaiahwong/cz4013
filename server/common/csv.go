package common

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
)

type CSVParser interface {
	Parse(data []string) error
}

func LoadCSV(file string, list interface{}) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	// Join the directory and file name
	absPath := filepath.Join(dir, file)

	csvFile, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	// Create a new CSV reader
	reader := csv.NewReader(csvFile)
	// Skip the first line
	_, err = reader.Read()
	if err != nil {
		return err
	}

	l := reflect.ValueOf(list)

	if l.Kind() != reflect.Ptr {
		return errors.New("list must be a pointer")
	}

	l = reflect.Indirect(l)

	for {
		// Read each record from CSV
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		t := reflect.Indirect(reflect.ValueOf(list)).Type()
		e := reflect.New(t.Elem().Elem())
		parsable, ok := e.Interface().(CSVParser)
		if !ok {
			return errors.New("Not a CSVParser")
		}
		parsable.Parse(record)
		// Append parsable to list
		l.Set(reflect.Append(l, e))
	}

	return nil
}
