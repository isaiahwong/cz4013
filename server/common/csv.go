package common

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
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

	paths := []string{
		// filepath.Join(dir, file),
		filepath.Join(dir, fmt.Sprintf("/server/%v", file)),
		filepath.Join(dir, fmt.Sprintf("../%v", file)),
	}
	var rerr error
	var csvFile *os.File

	// read csv fn
	readCsv := func(path string) (*os.File, error) {
		return os.Open(path)
	}

	stdInCsv := func() (*os.File, error) {
		scanner := bufio.NewScanner(os.Stdin)
		// Read input from standard input
		fmt.Print("\nEnter flight.csv location: ")
		scanner.Scan()
		path := scanner.Text()
		return readCsv(path)
	}

	for _, path := range paths {
		// Join the directory and file name
		csvFile, rerr = readCsv(path)
		if rerr == nil {
			break
		}
	}

	if rerr != nil {
		fmt.Println("\nAttempted to open file from these directories: ")
		for _, path := range paths {
			fmt.Println(path)
		}
		csvFile, rerr = stdInCsv()
		if rerr != nil {
			return rerr
		}
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
