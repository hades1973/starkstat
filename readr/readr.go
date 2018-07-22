// package readr read csv file or table file into Frame
package readr

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Frame struct {
	Dates   []string
	Opens   []float64
	Highs   []float64
	Closes  []float64
	Lows    []float64
	Volumns []float64
	Power   []float64
}

func ReadCSV(fname string, head bool) *Frame {
	f, err := os.Open(fname)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer f.Close()
	reader := csv.NewReader(f)
	all, err := reader.ReadAll()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if len(all) == 0 {
		return nil
	}
	if head && len(all) == 1 {
		return nil
	}
	if head {
		all = all[1:]
	}

	size := len(all)
	frame := Frame{
		Dates:   make([]string, size, size),
		Opens:   make([]float64, size, size),
		Highs:   make([]float64, size, size),
		Lows:    make([]float64, size, size),
		Closes:  make([]float64, size, size),
		Volumns: make([]float64, size, size),
		Power:   make([]float64, size, size),
	}
	for i, record := range all {
		frame.Dates[i] = record[0]
		frame.Opens[i], _ = strconv.ParseFloat(record[1], 64)
		frame.Highs[i], _ = strconv.ParseFloat(record[2], 64)
		frame.Lows[i], _ = strconv.ParseFloat(record[3], 64)
		frame.Closes[i], _ = strconv.ParseFloat(record[4], 64)
		frame.Volumns[i], _ = strconv.ParseFloat(record[5], 64)
		frame.Power[i], _ = strconv.ParseFloat(record[6], 64)
	}
	return &frame
}

func ReadTable(fname string, head bool) *Frame {
	f, err := os.Open(fname)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	frame := Frame{}
	if head {
		if scanner.Scan() != true {
			return nil
		}
	}
	for scanner.Scan() {
		record := strings.Fields(scanner.Text())
		frame.Dates = append(frame.Dates, record[0])
		var vs [6]float64
		for i := range record[1:] {
			vs[i], _ = strconv.ParseFloat(strings.TrimSpace(record[i]), 64)
		}
		frame.Opens = append(frame.Opens, vs[0])
		frame.Highs = append(frame.Highs, vs[1])
		frame.Closes = append(frame.Closes, vs[2])
		frame.Lows = append(frame.Lows, vs[3])
		frame.Volumns = append(frame.Volumns, vs[4])
		frame.Power = append(frame.Power, vs[5])
	}
	return &frame
}
