package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gocarina/gocsv"
)

func WriteToFile(records []PaymentRecord) (*os.File, error) {
	f, err := ioutil.TempFile("./records", "record")
	if err != nil {
		return nil, fmt.Errorf("cannot create temp file: %v", err)
	}

	err = gocsv.MarshalFile(records, f)
	if err != nil {
		return nil, fmt.Errorf("cannot write csv to the temp file: %v", err)
	}

	return f, nil
}
