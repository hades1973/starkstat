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
	"math"
	//	"net/http"
	"os"
	"path"
	"time"
)

/////////////////////////////////////////////////////////////////////
type CorrelationMatrix struct {
	m          [][]float64 // 相关系数矩阵
	StockCodes []string    // stockcodes
	stocks     int         // len(stockCodes)
}

func (c *CorrelationMatrix) Init(stocksCodes []string) {
	c.StockCodes = stocksCodes
	c.stocks = len(stocksCodes)
	c.m = make([][]float64, c.stocks, c.stocks)
	for i := 0; i < c.stocks; i++ {
		c.m[i] = make([]float64, c.stocks, c.stocks)
	}
	c.Unity()
}

func (c *CorrelationMatrix) Unity() {
	for i := 0; i < c.stocks; i++ {
		c.m[i][i] = 1.0
	}
}

func (c *CorrelationMatrix) At(i, j int) float64 {
	return c.m[i][j]
}

func (c *CorrelationMatrix) SetAt(i, j int, x float64) {
	c.m[i][j] = x
}

func (c *CorrelationMatrix) WriteMaxCorr(filename string) {
	// 将每个股票最大相关的股票打印出来

	if c.m == nil {
		fmt.Println("c.m is nil")
		return
	}
	if c.m[0] == nil {
		fmt.Println("c.m[0] is nil")
		return
	}

	// 寻找最大相关项，并记录
	var records []string
	for i := 0; i < c.stocks; i++ {
		max := func(xs []float64, forj int) (location int, value float64) {
			maxv, maxi := xs[0], 0 // 迄今为止xs最大元素
			for i := 1; i < len(xs); i++ {
				if forj == i { //(forj, forj)的相关系数为1, 不是我们所要的
					continue
				}
				if maxv < xs[i] {
					maxv = xs[i]
					maxi = i
				}
			}
			return maxi, maxv
		}
		location, value := max(c.m[i], i)
		records = append(records, fmt.Sprintf("%s,%s,%.4f", c.StockCodes[i], c.StockCodes[location], value))
	}

	// 将记录写入文件
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer f.Close()
	for _, rec := range records {
		fmt.Fprintln(f, rec)
	}
}

func (c *CorrelationMatrix) ComputeFor(I, J int) {
	// log for debug
	// fmt.Printf("computer for (%d, %d)\n", I, J)

	// sotckcodes[i], [j] 的相关性计算

	// 数据清洗，将数据日期一一对应，删除无法比对的数据
	dates, closesI, closesJ := LoadAndCleanData(c.StockCodes[I], c.StockCodes[J])
	if dates == nil || len(dates) < 5 {
		return
	}

	// 以一周为一个计算单位，计算相关性的频率
	counts := 0
	freqs := 0
	for {
		// 如果一周天数不足5天，为无效周，快进到下一周
		i_monday := NextMondayIndex(dates)
		if i_monday < 0 {
			break
		}
		dates = dates[i_monday:]
		closesI = closesI[i_monday:]
		closesJ = closesJ[i_monday:]
		i_monday = 0
		j_friday := NextFridayIndex(dates)
		if j_friday < 0 {
			break
		}
		if (j_friday - i_monday) != 4 { // 不在同一周
			dates = dates[j_friday:]
			closesI = closesI[j_friday:]
			closesJ = closesJ[j_friday:]
			continue
		}

		// 有效周，可以记录counts
		counts += 1

		// 一周股价
		cI := closesI[i_monday : j_friday+1]
		cJ := closesJ[i_monday : j_friday+1]
		dates = dates[j_friday:]
		closesI = closesI[j_friday:]
		closesJ = closesJ[j_friday:]

		// 标准化cI, cJ
		sumI, sumJ := 0.0, 0.0
		for i := 0; i < 5; i++ {
			sumI += cI[i] * cI[i]
			sumJ += cJ[i] * cJ[i]
		}
		sumI, sumJ = math.Sqrt(sumI), math.Sqrt(sumJ)
		for i := 0; i < 5; i++ {
			cI[i] /= sumI
			cJ[i] /= sumJ
		}

		// cI, cJ 均值，偏差
		sumI, sumJ = 0.0, 0.0
		for i := 0; i < 5; i++ {
			sumI += cI[i]
			sumJ += cJ[i]
		}
		meanI := sumI / 5.0
		meanJ := sumJ / 5.0
		diffI := make([]float64, 5, 5)
		diffJ := make([]float64, 5, 5)
		for i := 0; i < 5; i++ {
			diffI[i] = cI[i] - meanI
			diffJ[i] = cJ[i] - meanJ
		}

		// cI, cJ 协方差、标准差
		covar := 0.0
		for i := 0; i < 5; i++ {
			covar += diffI[i] * diffJ[i]
		}
		covar /= 4.0
		stdevI, stdevJ := 0.0, 0.0
		for i := 0; i < 5; i++ {
			stdevI += diffI[i] * diffI[i]
			stdevJ += diffJ[i] * diffJ[i]
		}
		stdevI /= (5.0 - 1.0)
		stdevI = math.Sqrt(stdevI)
		stdevJ /= (5.0 - 1.0)
		stdevJ = math.Sqrt(stdevJ)
		var corr = 0.0
		if stdevI > 0 && stdevJ > 0 {
			corr = covar / stdevI / stdevJ
		}
		if corr > 0.7 {
			freqs += 1
		}
	}
	if counts == 0 {
		c.m[I][J] = 0.0
	} else {
		c.m[I][J] = float64(freqs) / float64(counts)
	}
}

type CorrResult struct {
	i, j int
	corr float64
}

var limit = make(chan struct{}, 5)

func (c *CorrelationMatrix) MaxCorrFor(i int, corr chan<- CorrResult) {
	limit <- struct{}{}
	defer func() {
		<-limit
	}()

	for j := i + 1; j < c.stocks; j++ {
		c.ComputeFor(i, j)
	}
	// find max correlation and its position
	var maxv float64
	var maxj int
	xs := c.m[i]
	maxj = i + 1
	maxv = xs[i+1]
	for j := i + 1; j < c.stocks; j++ {
		if maxv < xs[j] {
			maxj = j
			maxv = xs[j]
		}
	}

	corr <- CorrResult{i, maxj, maxv}
}

func (c *CorrelationMatrix) WriteToCSV(filename string) {
	if c.m == nil {
		fmt.Println("c.m is nil")
		return
	}
	if c.m[0] == nil {
		fmt.Println("c.m[i] is nil")
		return
	}
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer f.Close()
	wr := csv.NewWriter(f)
	wr.Comma = ','
	wr.Write(append([]string{""}, c.StockCodes...))
	toStrings := func(xs []float64) []string {
		re := make([]string, len(xs))
		for i, x := range xs {
			re[i] = fmt.Sprintf("%.4f", x)
		}
		return re
	}

	for i := 0; i < len(c.StockCodes); i++ {
		wr.Write(append([]string{c.StockCodes[i]}, toStrings(c.m[i])...))
	}
	wr.Flush()
}

func LoadStockClose(stockcode string) (dates []string, closes []float64) {
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
		return nil, nil
	}
	defer f.Close()
	rd = csv.NewReader(f)
	rd.Comma = ','
	rd.Comment = '#'
	for {
		record, err = rd.Read()
		if err != nil {
			break
		}
		dates = append(dates, record[0])
		fmt.Sscanf(record[3], "%f", &close)
		closes = append(closes, close)
	}
	return
}

func LoadAndCleanData(f1name, f2name string) (dates []string, closes1, closes2 []float64) {
	datesI, closesI := LoadStockClose(f1name)
	datesJ, closesJ := LoadStockClose(f2name)
	// 数据清洗，将数据日期一一对应，删除无法比对的数据
	var (
		i, j int
	)
	i, j = 0, 0
	for {
		if i >= len(datesI) || j >= len(datesJ) {
			break
		}

		if datesI[i] < datesJ[j] {
			i++
			continue
		} else if datesI[i] > datesJ[j] {
			j++
			continue
		}
		dates = append(dates, datesI[i])
		closes1 = append(closes1, closesI[i])
		closes2 = append(closes2, closesJ[j])
		i++
		j++
	}
	return
}

const timeLayout = "2006-01-02"

func NextMondayIndex(dates []string) int {
	for i, date := range dates {
		t, _ := time.Parse(timeLayout, date)
		if t.Weekday() == time.Monday {
			return i
		}
	}
	return -1
}

func NextFridayIndex(dates []string) int {
	for i, date := range dates {
		t, _ := time.Parse(timeLayout, date)
		if t.Weekday() == time.Friday {
			return i
		}
	}
	return -1
}
