// go build fetchAll.go
package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"
)

//////////////////////////////////////////////////////////
// main

const (
	DIR = "/home/jns/diskD/stockdata/"
)

var StockListFileName = path.Join(DIR, "stocklist.csv")

func main() {
	// 从stocklist.csv读入股票代码、名称等.
	f, err := os.Open(StockListFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	reader := csv.NewReader(f)
	reader.Comma = ','
	reader.TrimLeadingSpace = true

	// 读取股票列表文件，并行爬取数据
	done := make(chan struct{})
	var wg sync.WaitGroup // 记录活动的go routines
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			return
		}
		wg.Add(1)
		go func(stockCode, stockName string) {
			defer wg.Done()
			ModifyStockData(stockCode, stockName)
			done <- struct{}{}
			fmt.Println(stockName, "modified")
		}(record[0], record[1])

	}
	// 监视 go routines 返回, 然后关闭 done channel
	go func() {
		wg.Wait()
		close(done)
	}()
	for range done {
	}
}

/////////////////////////////////////////////////////////
// ModifyStockData
////////////////////////////////////////////////////////

var (
	limitedThreads = make(chan struct{}, 5)
)

func ModifyStockData(stockcode, stockname string) {
	limitedThreads <- struct{}{}
	defer func() {
		<-limitedThreads
	}()

	// f for 股票数据文件, doc 股票数据原文件内容
	// dates for recording stock deal date,
	// deals for recording stock open,high,close,low,volumn,transaction,pow
	// csvRd for reading data from f
	// csvWr for writing data into f
	var (
		f     *os.File
		doc   []byte
		err   error
		dates []string
		deals [][7]float64
		csvRd *csv.Reader
		csvWr *csv.Writer
	)

	// 读出股票数据, 并将价格除权
	doc, err = ioutil.ReadFile(path.Join(DIR, stockcode+".csv"))
	if err != nil {
		fmt.Println("Can't read ", stockname)
		return
	}
	rd := bufio.NewReader(bytes.NewReader(doc))
	_, _, err = rd.ReadLine() // skip first line(header)
	if err != nil {
		return
	}
	csvRd = csv.NewReader(rd)
	csvRd.FieldsPerRecord = 8
	csvRd.Comma = ','
	var deal [7]float64 // for open,high,close,low,volumn,transaction,pow.
	for {
		record, err := csvRd.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		date := record[0]
		record = record[1:]
		for i := 0; i < len(record); i++ {
			fmt.Sscanf(record[i], "%f", &deal[i])
		}
		for i := 0; i < 4; i++ { // 除权后的股价
			deal[i] /= deal[6]
		}
		dates = append(dates, date)
		deals = append(deals, deal)
	}

	// 将改变后的数据重新写入股票数据文件
	f, err = os.Create(path.Join(DIR, stockcode+".csv"))
	if err != nil {
		fmt.Println("cant create ", stockname)
		return
	}
	defer f.Close()
	csvWr = csv.NewWriter(f)
	csvWr.Comma = ','
	csvWr.UseCRLF = true
	comment := []string{"#date", "open", "high", "cloe", "low", "volumn", "pow"}
	csvWr.Write(comment)
	var record [7]string
	for i, date := range dates {
		//deal[0-7]: "#date", "open", "high", "cloe", "low", "volumn", "transaction", "pow"
		deal := deals[i]
		record[0] = date
		rec := record[1:] // for convient modiy record
		for j := 0; j < 4; j++ {
			rec[j] = fmt.Sprintf("%.3f", deal[j]) //use rec, same j as deal, but purpose is modify record
		}
		rec[4] = fmt.Sprintf("%.0f", deal[4])
		rec[5] = fmt.Sprintf("%.3f", deal[6]) // discard "transaction"
		csvWr.Write(record[:])
	}
	csvWr.Flush()
}
