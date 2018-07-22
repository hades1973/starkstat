// 本程序用来修改原有的stocklist.csv文件。
// 将字段"stockLastUpdateDay"以及"lastPower"值加入文件。

package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path"
)

//////////////////////////////////////////////////////////
// main

const (
	DIR               = "/home/jns/diskD/stockdata/"
	StockListFileName = "stocklist.csv"
)

func main() {
	// encoding/csv 解码输入文件
	var (
		f   *os.File
		err error
		all [][]string
	)
	f, err = os.Open(path.Join(DIR, StockListFileName))
	if err != nil {
		log.Fatal(err)
	}
	reader := csv.NewReader(f)
	reader.Comma = ','
	reader.FieldsPerRecord = 2
	reader.TrimLeadingSpace = true
	all, err = reader.ReadAll()
	if err != nil {
		f.Close()
		log.Fatal(err)
		return
	}
	f.Close()

	// 重新打开stocklist文件，准备写入
	f, err = os.Create(path.Join(DIR, StockListFileName))
	if err != nil {
		log.Fatal(err)
		return
	}
	defer f.Close()
	// 逐个读取log文件，解析出"date", "power"字段
	var (
		newRecord = make(chan string)
		i         int
		record    []string
	)
	for i, record = range all {
		go func(stockCode, stockName string) {
			ModifyStockList(stockCode, stockName, newRecord)
			fmt.Println(stockName, "modified")
		}(record[0], record[1])
	}
	for i > 0 {
		i = i - 1

		fmt.Fprintf(f, "%s", <-newRecord)
	}
}

/////////////////////////////////////////////////////////
// ModifyStockData
////////////////////////////////////////////////////////

var (
	limitedThreads = make(chan struct{}, 5)
)

func ModifyStockList(stockcode, stockname string, newRecord chan<- string) {
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
		doc   []string
		err   error
		rd    *csv.Reader
		date  string
		power string
	)

	// 读出股票log, 解析出日期与权值
	f, err = os.Open(path.Join(DIR, stockcode+".log"))
	if err != nil {
		log.Fatal(err)
		return
	}
	rd = csv.NewReader(f)
	rd.Comma = ','
	rd.TrimLeadingSpace = true
	doc, err = rd.Read()
	if err != nil {
		f.Close()
		log.Fatal()
		return
	}
	f.Close()
	date, power = doc[0], doc[7]

	// 写入channel newRecord
	newRecord <- fmt.Sprintf("%s,%s,%s,%s\r\n", stockcode, stockname, date, power)
	return
}
