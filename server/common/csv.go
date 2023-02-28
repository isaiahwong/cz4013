package common

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
	"reflect"
)

type CSVParser interface {
	Parse(data []string) error
}

func LoadCSV(file string, list interface{}) error {
	csvFile, err := os.Open(file)
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
		e := reflect.New(t.Elem())
		parsable, ok := e.Interface().(CSVParser)
		if !ok {
			return errors.New("Not a CSVParser")
		}
		parsable.Parse(record)
		// Append parsable to list
		reflect.ValueOf(list).Elem().Set(
			reflect.Append(
				reflect.ValueOf(list).Elem(), e.Elem(),
			),
		)

	}

	return nil
}
