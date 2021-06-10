[![GoDoc](https://godoc.org/github.com/cinar/csv2?status.svg)](https://godoc.org/github.com/cinar/csv2)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://travis-ci.com/cinar/csv2.svg?branch=master)](https://travis-ci.com/cinar/csv2)

# Csv2 Go

Csv2 is a lightweight Golang module for reading CSV files as individual rows or as a table.

## Usage

Install package.

```bash
go get github.com/cinar/csv2
```

Import Csv2.

```Golang
import (
    "github.com/cinar/csv2"
)
```

### Reading as individual rows

Given that the CSV file contains the following columns.

```CSV
date,close,high,low,open,volume,adjClose,adjHigh,adjLow,adjOpen,adjVolume,divCash,splitFactor
2015-09-18 00:00:00+00:00,43.48,43.99,43.33,43.5,63143684,39.5167038561,39.9802162518,39.3803766809,39.534880812800004,63143684,0.0,1.0
```

Define a structure for each individual row.

```Golang
// Daily price structure for each row.
type dailyPrice struct {
	Date        time.Time `format:"2006-01-02 15:04:05-07:00"`
	Close       float64
	High        float64
	Low         float64
	Open        float64
	Volume      int64
	AdjClose    float64
	AdjHigh     float64
	AdjLow      float64
	AdjOpen     float64
	AdjVolume   int64
	DivCash     float64
	SplitFactor float64
}
```

Csv2 allows you to associate additional information about the colums through the tags. The following additional information is currently supported.

Tag | Description | Example
--- | --- | ---
header | Column header for the field. | `header:"Date"`
format | Date format for parsing. | `format:"2006-01-02 15:04:05-07:00"`

Define an instance of a slice of row structure.

```Golang
var prices []dailyPrice
```

Use the [ReadRowsFromFile](https://pkg.go.dev/github.com/cinar/csv2#ReadRowsFromFile) function to read the CSV file into the slice.

```Golang
err := ReadRowsFromFile(testFile, true, &prices)
if err != nil {
    return err
}
```
## License

The source code is provided under MIT License.
