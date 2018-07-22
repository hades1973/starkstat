package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path"
	"sort"
	"stockstat/readr"
)

type StatResult struct {
	StockCode            string
	Ok                   bool // 记录是否是有效StatResult
	Days                 []int
	BeginDate, EndDate   []string
	BeginPrice, EndPrice []float64
	DeltaPrice           []float64
	MeanDelta            []float64
}

type LastRecord struct {
	StockCode      string
	LastDeltaPrice float64
	Mean           float64
}

type LastRecordSeq []LastRecord

func (lrec LastRecordSeq) Len() int {
	return len(lrec)
}

func (lrec LastRecordSeq) Less(i, j int) bool {
	return (lrec[i].Mean < lrec[j].Mean)
}

func (lrec LastRecordSeq) Swap(i, j int) {
	lrec[i], lrec[j] = lrec[j], lrec[i]
}

const DIR = "/home/jns/diskD/stockdata/"
const STOCKROSTER = "stocklist.csv"

// stockmap map stockcode to stockname
var stockmap = make(map[string]string)
var fo = os.Stdout

func main() {
	// f for read stock's code and name
	f, err := os.Open(path.Join(DIR, STOCKROSTER))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	reader := csv.NewReader(f)

	// f2 for write result
	f2, err := os.Create("sort.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f2.Close()

	// 打印表头
	fmt.Fprintf(fo, "%s\r\n", "StockCode,BeginDate,EndDate,Days,BeginPrice,EndPrice,DeltaPrice,MeanDelta")
	n := 0                           // 记录gouroutine数量
	sts := make(chan *StatResult, 0) // 统计结果

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			return
		}
		stockcode := record[0]
		stockname := record[1]
		stockmap[stockcode] = stockname
		n++
		go Statistic(stockcode, sts)
	}
	lastRecords := make([]LastRecord, 0) // 记录最后一个DelatPrice以备排序
	for n > 0 {
		// print result
		res := <-sts
		len := len(res.Days)
		mean := 0.0
		i := 0
		for ; i < len; i++ {
			mean += res.MeanDelta[i]
			fmt.Fprintf(fo, "%s,%s,%s,%d,%f,%f,%f,%f\r\n",
				res.StockCode, res.BeginDate[i], res.EndDate[i],
				res.Days[i],
				res.BeginPrice[i], res.EndPrice[i], res.DeltaPrice[i], res.MeanDelta[i])
		}
		if i > 8 && res.EndPrice[i-1] < 10.0 && res.DeltaPrice[i-2] < -30.0 && res.DeltaPrice[i-1] < 10.0 {
			lastRecords = append(lastRecords, LastRecord{StockCode: res.StockCode, LastDeltaPrice: res.DeltaPrice[i-2], Mean: mean / float64(i)})
		}

		n--
	}

	sort.Sort(LastRecordSeq(lastRecords))
	fmt.Fprintf(f2, "%s,%s,%s\r\n", "stockcode", "LastDeltaPrice", "Mean")
	for j := len(lastRecords) - 1; j >= 0; j-- {
		fmt.Fprintf(f2, "%s,%f,%f\r\n", lastRecords[j].StockCode, lastRecords[j].LastDeltaPrice, lastRecords[j].Mean)
	}
}

// limitedRoutines limit the concurrent Staticstic routines.
var limitedRoutines = make(chan struct{}, 4)

func Statistic(code string, sts chan *StatResult) {
	limitedRoutines <- struct{}{}
	defer func() {
		<-limitedRoutines
	}()

	res := StatResult{StockCode: code, Ok: true}
	defer func() {
		sts <- &res
	}()
	dat := readr.ReadCSV(path.Join(DIR, code+".csv"), true)
	if dat == nil {
		res.Ok = false
		return
	}
	//	for i := range dat.Closes {
	//		dat.Closes[i] /= dat.Power[i]
	//	}

	j, k := 1, 0
	for ; j < len(dat.Dates); j = j + 1 {
		// k record begin index for same power
		vj1, vj0 := dat.Power[j], dat.Power[j-1]
		if vj1 != vj0 && math.Abs(vj1-vj0)/vj0 > 0.005 {
			res.BeginDate = append(res.BeginDate, dat.Dates[k])
			res.BeginPrice = append(res.BeginPrice, dat.Closes[k])
			res.EndDate = append(res.EndDate, dat.Dates[j-1])
			res.EndPrice = append(res.EndPrice, dat.Closes[j-1])
			res.Days = append(res.Days, j-k)
			deltap := (dat.Closes[j-1] - dat.Closes[k])
			res.DeltaPrice = append(res.DeltaPrice, deltap/dat.Closes[k]*100.0)
			res.MeanDelta = append(res.MeanDelta, deltap/float64(j-k)/dat.Closes[k]*100.0)

			k = j
		}
	}
	res.BeginDate = append(res.BeginDate, dat.Dates[k])
	res.BeginPrice = append(res.BeginPrice, dat.Closes[k])
	res.EndDate = append(res.EndDate, dat.Dates[j-1])
	res.EndPrice = append(res.EndPrice, dat.Closes[j-1])
	res.Days = append(res.Days, j-k)
	deltap := (dat.Closes[j-1] - dat.Closes[k])
	res.DeltaPrice = append(res.DeltaPrice, deltap/dat.Closes[k]*100.0)
	res.MeanDelta = append(res.MeanDelta, deltap/float64(j-k)/dat.Closes[k]*100.0)
}
