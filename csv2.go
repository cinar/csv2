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

func setBoolFieldValue(fieldValue reflect.Value, stringValue string) error {
	value, err := strconv.ParseBool(stringValue)
	if err == nil {
		fieldValue.SetBool(value)
	}

	return err
}

func setIntFieldValue(fieldValue reflect.Value, stringValue string, bitSize int) error {
	value, err := strconv.ParseInt(stringValue, 10, bitSize)
	if err == nil {
		fieldValue.SetInt(value)
	}

	return err
}

func setUintFieldValue(fieldValue reflect.Value, stringValue string, bitSize int) error {
	value, err := strconv.ParseUint(stringValue, 10, bitSize)
	if err == nil {
		fieldValue.SetUint(value)
	}

	return err
}

func setFloatFieldValue(fieldValue reflect.Value, stringValue string, bitSize int) error {
	value, err := strconv.ParseFloat(stringValue, bitSize)
	if err == nil {
		fieldValue.SetFloat(value)
	}

	return err
}

func setTimeFieldValue(fieldValue reflect.Value, stringValue string, format string) error {
	value, err := time.Parse(format, stringValue)
	if err == nil {
		fieldValue.Set(reflect.ValueOf(value))
	}

	return err
}

func setFieldValue(fieldValue reflect.Value, stringValue string, format string) error {
	fieldKind := fieldValue.Kind()

	switch fieldKind {
	case reflect.String:
		fieldValue.SetString(stringValue)
		return nil

	case reflect.Bool:
		return setBoolFieldValue(fieldValue, stringValue)

	case reflect.Int:
		return setIntFieldValue(fieldValue, stringValue, bits.UintSize)

	case reflect.Int8:
		return setIntFieldValue(fieldValue, stringValue, 8)

	case reflect.Int16:
		return setIntFieldValue(fieldValue, stringValue, 16)

	case reflect.Int32:
		return setIntFieldValue(fieldValue, stringValue, 32)

	case reflect.Int64:
		return setIntFieldValue(fieldValue, stringValue, 64)

	case reflect.Uint:
		return setUintFieldValue(fieldValue, stringValue, bits.UintSize)

	case reflect.Uint8:
		return setUintFieldValue(fieldValue, stringValue, 8)

	case reflect.Uint16:
		return setUintFieldValue(fieldValue, stringValue, 16)

	case reflect.Uint32:
		return setUintFieldValue(fieldValue, stringValue, 32)

	case reflect.Uint64:
		return setUintFieldValue(fieldValue, stringValue, 64)

	case reflect.Float32:
		return setFloatFieldValue(fieldValue, stringValue, 32)

	case reflect.Float64:
		return setFloatFieldValue(fieldValue, stringValue, 64)

	case reflect.Struct:
		fieldTypeString := fieldValue.Type().String()

		switch fieldTypeString {
		case "time.Time":
			return setTimeFieldValue(fieldValue, stringValue, format)

		default:
			return fmt.Errorf("unsupported struct type %s", fieldTypeString)
		}

	default:
		return fmt.Errorf("unsupported field kind %s", fieldKind)
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
			if err = setFieldValue(row.Field(column.FieldIndex), record[column.ColumnIndex], column.Format); err != nil {
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
