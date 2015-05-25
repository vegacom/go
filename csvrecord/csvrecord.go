/*
Package csvrecord reads a CSV file into a a slice of Record. A Record represents a row in the CSV,
with each key corresponding to a column header from the first line.
*/
package csvrecord

// TODO: add unit tests.

import (
	"encoding/csv"
	"errors"
	"io"
)

type Record map[string]string

func Read(r io.Reader) ([]Record, error) {
	reader := csv.NewReader(r)
	reader.TrailingComma = true
	reader.LazyQuotes = true
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}
	var records []Record
	for {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if len(row) != len(header) {
			return nil, errors.New("csv contains rows with different row numbers")
		}
		rec := make(map[string]string)
		for i, key := range header {
			rec[key] = row[i]
		}
		records = append(records, rec)
	}
	return records, nil
}
