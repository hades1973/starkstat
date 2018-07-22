package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"readr"
	"strings"
)

func FileList(path string) ([]string, error) {
	fs := []string{}
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		if strings.Index(f.Name(), ".") == -1 {
			fs = append(fs, f.Name())
		}
		return nil
	})
	return fs, err
}

func main() {
	root := "./"
	if len(os.Args) == 2 {
		root = os.Args[1]
	}
	fs, err := FileList(root)
	if err != nil {
		fmt.Println(err)
	}
	for _, fname := range fs {
		frm := readr.ReadTable(path.Join(root, fname), false)
		if frm == nil {
			continue
		}
		if f, err := os.Create(path.Join(root, fname+".csv")); err == nil {
			fmt.Fprintf(f, "date,open,high,close,low,volumn,transaction,power\n")
			for i := range frm.Dates {
				fmt.Fprintf(f, "%s,%f,%f,%f,%f,%f,%f,%f\n",
					frm.Dates[i], frm.Opens[i], frm.Highs[i], frm.Closes[i], frm.Lows[i], frm.Volumns[i], frm.Transactions[i], frm.Power[i])
			}
			f.Close()
			os.Remove(path.Join(root, fname))
		}

	}
}
