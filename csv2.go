package csv2

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math/bits"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	// Header name
	TagHeader = "header"

	// Format name
	TagFormat = "format"
)

const (
	timeFormat = "2006-01-02 15:04:05"
)

type columnInfo struct {
	Header      string
	ColumnIndex int
	FieldIndex  int
	Format      string
}

func setBoolValue(value reflect.Value, stringValue string) error {
	actualValue, err := strconv.ParseBool(stringValue)
	if err == nil {
		value.SetBool(actualValue)
	}

	return err
}

func setIntValue(value reflect.Value, stringValue string, bitSize int) error {
	actualValue, err := strconv.ParseInt(stringValue, 10, bitSize)
	if err == nil {
		value.SetInt(actualValue)
	}

	return err
}

func setUintValue(value reflect.Value, stringValue string, bitSize int) error {
	actualValue, err := strconv.ParseUint(stringValue, 10, bitSize)
	if err == nil {
		value.SetUint(actualValue)
	}

	return err
}

func setFloatValue(value reflect.Value, stringValue string, bitSize int) error {
	actualValue, err := strconv.ParseFloat(stringValue, bitSize)
	if err == nil {
		value.SetFloat(actualValue)
	}

	return err
}

func setTimeValue(value reflect.Value, stringValue string, format string) error {
	actualValue, err := time.Parse(format, stringValue)
	if err == nil {
		value.Set(reflect.ValueOf(actualValue))
	}

	return err
}

func setValue(value reflect.Value, stringValue string, format string) error {
	kind := value.Kind()

	switch kind {
	case reflect.String:
		value.SetString(stringValue)
		return nil

	case reflect.Bool:
		return setBoolValue(value, stringValue)

	case reflect.Int:
		return setIntValue(value, stringValue, bits.UintSize)

	case reflect.Int8:
		return setIntValue(value, stringValue, 8)

	case reflect.Int16:
		return setIntValue(value, stringValue, 16)

	case reflect.Int32:
		return setIntValue(value, stringValue, 32)

	case reflect.Int64:
		return setIntValue(value, stringValue, 64)

	case reflect.Uint:
		return setUintValue(value, stringValue, bits.UintSize)

	case reflect.Uint8:
		return setUintValue(value, stringValue, 8)

	case reflect.Uint16:
		return setUintValue(value, stringValue, 16)

	case reflect.Uint32:
		return setUintValue(value, stringValue, 32)

	case reflect.Uint64:
		return setUintValue(value, stringValue, 64)

	case reflect.Float32:
		return setFloatValue(value, stringValue, 32)

	case reflect.Float64:
		return setFloatValue(value, stringValue, 64)

	case reflect.Struct:
		typeString := value.Type().String()

		switch typeString {
		case "time.Time":
			return setTimeValue(value, stringValue, format)

		default:
			return fmt.Errorf("unsupported struct type %s", typeString)
		}

	default:
		return fmt.Errorf("unsupported value kind %s", kind)
	}
}

func getStructFieldsAsColumns(structType reflect.Type) []columnInfo {
	columns := make([]columnInfo, structType.NumField())
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		header, ok := field.Tag.Lookup(TagHeader)
		if !ok {
			header = field.Name
		}

		format, ok := field.Tag.Lookup(TagFormat)
		if !ok {
			format = timeFormat
		}

		columns[i] = columnInfo{
			Header:      header,
			ColumnIndex: i,
			FieldIndex:  i,
			Format:      format,
		}
	}

	return columns
}

func readHeader(csvReader csv.Reader, columns []columnInfo) error {
	headers, err := csvReader.Read()
	if err != nil {
		return err
	}

	for _, column := range columns {
		for i, header := range headers {
			if strings.EqualFold(column.Header, header) {
				column.ColumnIndex = i
				break
			}
		}
	}

	return nil
}

// Read rows from reader.
func ReadRowsFromReader(reader io.Reader, hasHeader bool, rows interface{}) error {
	rowsPtrType := reflect.TypeOf(rows)
	if rowsPtrType.Kind() != reflect.Ptr {
		return errors.New("rows not a pointer")
	}

	rowsSliceType := rowsPtrType.Elem()
	if rowsSliceType.Kind() != reflect.Slice {
		return errors.New("rows not a pointer to slice")
	}

	rowType := rowsSliceType.Elem()
	if rowType.Kind() != reflect.Struct {
		return errors.New("rows not a pointer to slice of struct")
	}

	rowsPtr := reflect.ValueOf(rows)
	rowsSlice := rowsPtr.Elem()

	columns := getStructFieldsAsColumns(rowType)

	csvReader := csv.NewReader(reader)

	if hasHeader {
		if err := readHeader(*csvReader, columns); err != nil {
			return err
		}
	}

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		row := reflect.New(rowType).Elem()

		for _, column := range columns {
			if err = setValue(row.Field(column.FieldIndex), record[column.ColumnIndex], column.Format); err != nil {
				return err
			}
		}

		rowsSlice = reflect.Append(rowsSlice, row)
	}

	rowsPtr.Elem().Set(rowsSlice)

	return nil
}

// Read rows from file.
func ReadRowsFromFile(fileName string, hasHeader bool, rows interface{}) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}

	defer file.Close()

	return ReadRowsFromReader(file, hasHeader, rows)
}

// Read table from reader.
func ReadTableFromReader(reader io.Reader, hasHeader bool, table interface{}) error {
	tablePtrType := reflect.TypeOf(table)
	if tablePtrType.Kind() != reflect.Ptr {
		return errors.New("table not a pointer")
	}

	tableType := tablePtrType.Elem()
	if tableType.Kind() != reflect.Struct {
		return errors.New("table not a pointer to struct")
	}

	for i := 0; i < tableType.NumField(); i++ {
		if tableType.Field(i).Type.Kind() != reflect.Slice {
			return errors.New("table fields must be all slices")
		}
	}

	tableValue := reflect.ValueOf(table).Elem()

	columns := getStructFieldsAsColumns(tableType)

	csvReader := csv.NewReader(reader)

	if hasHeader {
		if err := readHeader(*csvReader, columns); err != nil {
			return err
		}
	}

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		for _, column := range columns {
			sliceValue := tableValue.Field(column.FieldIndex)

			itemValue := reflect.New(sliceValue.Type().Elem()).Elem()
			if err = setValue(itemValue, record[column.ColumnIndex], column.Format); err != nil {
				return err
			}

			sliceValue.Set(reflect.Append(sliceValue, itemValue))
		}
	}

	return nil
}

// Read table from file.
func ReadTableFromFile(fileName string, hasHeader bool, rows interface{}) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}

	defer file.Close()

	return ReadTableFromReader(file, hasHeader, rows)
}
