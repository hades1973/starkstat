// corr 根据股票列表文件 stocklist.csv 与 各只股票的价格文件
// 计算两两的相关性。
// 程序输出corr.csv文件，将股票相关系数矩阵以csv格式输出。
//			,stock1, stock2, ...
// stock1	,xi_11,  xi_12, ...
// stock2	,xi_21,  xi_22, ...

package main

import (
	"encoding/csv"
	"fmt"
	"log"
	//	"math"
	//	"net/http"
	"os"
	"path"
	//	"time"
)

const (
	DIR = "./stockdata/"
)

var (
	StockListFileName = path.Join(DIR, "stocklist.csv")
	ResultFileName    = "corr.csv"
	Corr              = new(CorrelationMatrix)
)

// 统计、计算相关性
var completePercent = 0.0

func statcorr() {
	f, err := os.Create("corr.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	stockcodes := GetStockCodes()
	if stockcodes == nil {
		fmt.Println("error, no stocklist found!")
		return
	}
	Corr.Init(stockcodes)

	var corr = make(chan CorrResult)
	first, end := 0, len(stockcodes)-2
	for {
		if first > end {
			break
		} else if first < end {
			go Corr.MaxCorrFor(first, corr)
			go Corr.MaxCorrFor(end, corr)
		} else {
			go Corr.MaxCorrFor(first, corr)
		}
		first++
		end--
	}

	var c CorrResult
	for i := 0; i < len(stockcodes)-1; i++ {
		c = <-corr
		fmt.Fprintf(f, "%d, %d, %.2f\n", Corr.StockCodes[c.i], Corr.StockCodes[c.j], c.corr)
		fmt.Printf("finished: %.2f %%\n", float64(i)/float64(Corr.stocks-1))
	}
}

func main() {
	statcorr()
}

func GetStockCodes() []string {
	var (
		f          *os.File
		err        error
		stockCodes []string = make([]string, 0)
	)
	f, err = os.Open(StockListFileName)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer f.Close()
	rd := csv.NewReader(f)
	rd.Comma = ','
	rd.Comment = '#'
	rd.TrimLeadingSpace = true
	all, err := rd.ReadAll()
	if err != nil {
		log.Fatal(err)
		return nil
	}
	for _, row := range all {
		stockCodes = append(stockCodes, row[0])
	}
	return stockCodes
}
