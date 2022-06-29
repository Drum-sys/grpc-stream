package main

import (
	"bufio"
	"fmt"
	"github.com/golang/snappy"
	"os"
	"time"
)

func ReadBlock2(filePath string) error  {
	//var wg sync.WaitGroup
	var count = 0
	var line string

	f, err := os.Open(filePath)
	fmt.Println("err", err)
	defer f.Close()
	if err != nil {
		return err
	}
	buf := bufio.NewScanner(f)
	for {
		if !buf.Scan() {
			break
		}
		line += buf.Text()
		count++
		if count == 200000 {
			fmt.Println("source len:", len([]byte(line)))
			t := time.Now()
			partCompress := snappy.Encode(nil, []byte(line))
			fmt.Println("readAll spend : ", time.Now().Sub(t))
			fmt.Println("compressed len:", len(partCompress))
			count = 0
			line = ""
		}
	}

	return nil
}

func main()  {
	ReadBlock2("part.txt")
}