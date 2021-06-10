package csv2

import (
	"testing"
	"time"
)

const (
	testFile = "test.csv"
)

// date,close,high,low,open,volume,adjClose,adjHigh,adjLow,adjOpen,adjVolume,divCash,splitFactor

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

// Stock prices structure for all columns.
type stockPrices struct {
	Date        []time.Time
	Close       []float64
	High        []float64
	Low         []float64
	Open        []float64
	Volume      []int64
	AdjClose    []float64
	AdjHigh     []float64
	AdjLow      []float64
	AdjOpen     []float64
	AdjVolume   []int64
	DivCash     []float64
	SplitFactor []float64
}

func TestReadRowsFromFile(t *testing.T) {
	var prices []dailyPrice

	err := ReadRowsFromFile(testFile, true, &prices)
	if err != nil {
		t.Fatal(err)
	}

	if n := len(prices); n != 10 {
		t.Fatalf("prices must have 10 element but has %d", n)
	}
}
