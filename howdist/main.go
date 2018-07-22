package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/gonum/stat"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: howdist stockcoe")
	}
	xs, dates := LoadStockClose(os.Args[1])
	for i := 0; i < len(dates)-1; i++ {
		y := xs[dates[i]:dates[i+1]]
		x := DeltaPrices(y)
		fmt.Printf("mean = %.1f, stddev=%.1f, numbers=%d\n", stat.Mean(x, nil), stat.StdDev(x, nil), len(x))
	}

}

// DElataPrices返回股价序列的涨跌值序列
func DeltaPrices(prices []float64) []float64 {
	if prices == nil || len(prices) < 2 {
		return nil
	}
	xs := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		xs = append(xs, (prices[i]-prices[i-1])/prices[i-1]*100.)
	}
	return xs
}

// LoadStockClose根据股票代码载入该股票的收盘价及对应的除权日
// closes[dates[i]]...closes[dates[i+1]之间的权值相同。
// dates[i]的值是closes的下标
func LoadStockClose(stockcode string) (closes []float64, dates []int) {
	const DIR = "/home/jns/diskD/stockdata/"
	var (
		fname  string
		f      *os.File
		rd     *csv.Reader
		err    error
		record []string
		close  float64
	)
	fname = path.Join(DIR, stockcode+".csv")
	f, err = os.Open(fname)
	if err != nil {
		log.Fatal(err)
		return nil, nil
	}
	defer f.Close()
	rd = csv.NewReader(f)
	rd.Comma = ','
	rd.Comment = '#'
	var i = 0
	pow := "no meaning"
	for {
		record, err = rd.Read()
		if err != nil {
			break
		}
		if record[6] != pow {
			pow = record[6]
			dates = append(dates, i)
		}
		fmt.Sscanf(record[3], "%f", &close)
		closes = append(closes, close)
		i++
	}
	dates = append(dates, i)
	return closes, dates
}
