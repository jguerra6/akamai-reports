package fileHandling

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/jguerra6/akamai-reports/accountSwitch"
)

const (
	dataBasePath      = "./data/master"
	switchKeyFileName = "switch_keys.json"
)

func SaveToFile(data interface{}, filePath string, fileName string, format string, delimiter ...rune) error {
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	fileName = filePath + "/" + fileName

	if format == "csv" {
		delimiterStr := delimiter
		_ = delimiterStr
		return saveToCSV(data, fileName, delimiter[0])
	}

	file, _ := json.MarshalIndent(data, "", " ")

	err := os.WriteFile(fileName, file, 0644)
	if err != nil {
		return err
	}

	return nil
}

// saveToCSV saves data from a struct slice to a CSV file
func saveToCSV(data interface{}, filename string, delimiter ...rune) error {
	// Open the CSV file for writing
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	if len(delimiter) == 1 {
		writer.Comma = delimiter[0]
	}

	// Use reflection to get the struct type
	structType := reflect.TypeOf(data).Elem()

	// Write the CSV header based on struct tags
	header := make([]string, 0)
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		csvTag := field.Tag.Get("csv")
		header = append(header, csvTag)
	}
	writer.Write(header)

	// Write the data rows to the CSV file
	structValue := reflect.ValueOf(data)
	for i := 0; i < structValue.Len(); i++ {
		row := make([]string, 0)
		for j := 0; j < structValue.Index(i).NumField(); j++ {
			fieldValue := structValue.Index(i).Field(j)
			row = append(row, fmt.Sprintf("%v", fieldValue.Interface()))
		}
		writer.Write(row)
	}

	return nil
}

func LoadSwitchKeys() ([]*accountSwitch.AccountSwitch, error) {
	fileName := dataBasePath + "/" + switchKeyFileName

	byteValue, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	var switchKeys []*accountSwitch.AccountSwitch

	err = json.Unmarshal(byteValue, &switchKeys)
	if err != nil {
		return nil, err
	}
	return switchKeys, nil
}
