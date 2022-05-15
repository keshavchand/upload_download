package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

type TestInterface interface {
  io.Writer
  Read(int) int
}
// const DownloadSize int64 = 100
const DownloadSize int64 = 1024 * 1024 * 25

type FileContents struct {
	offset  int64
	content []byte
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal(os.Args[0], " http://site filename")
	}
	site := os.Args[1]
	filename := os.Args[2]
	writer := make(chan FileContents, 10)
	go WriteToFile(filename, writer)
	resp, err := http.Head(site)
	if err != nil {
		log.Fatal(err)
	}

	length := resp.ContentLength
	var wg sync.WaitGroup
	defer wg.Wait()

	downloadLimiter := make(chan bool, 100)
	offset := int64(0)
	for ; offset*DownloadSize <= length; offset += 1 {
		wg.Add(1)
		downloadLimiter <- true
		go func(offset int64) {
			defer func() { _ = <-downloadLimiter }()
			defer wg.Done()
			req, err := http.NewRequest("GET", site, nil)
			if err != nil {
				fmt.Println(err)
				return
			}
			starting := offset * DownloadSize
			ending := min((offset+1)*DownloadSize-1, length)
			size := ending - starting
			fileRange := fmt.Sprintf("bytes=%d-%d", starting, ending)
			// log.Println(fileRange, size)
			req.Header.Add("Range", fileRange)
			client := &http.Client{}
			resp, err := client.Do(req)
			defer resp.Body.Close()
			if err != nil {
				fmt.Println(err)
				return
			}

			contents, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatalf(err) // this case should be handled
        return
			}
			var fc FileContents
			fc.offset = offset * DownloadSize
			fc.content = contents
			writer <- fc
		}(offset)
	}
}

func WriteToFile(filename string, fileContents <-chan FileContents) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
reader:
	for {
		fcont, ok := <-fileContents
		if !ok {
			break reader
		}
		log.Println(len(fcont.content))
		file.WriteAt(fcont.content, fcont.offset)
	}
}
func min(a, b int64) int64 {
	if a < b {
		return a
	}

	return b
}
