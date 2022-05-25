package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

// const DownloadSize int64 = 100
// const DownloadSize int64 = 2
const DownloadSize int64 = 1024 * 1024 * 10

type FileContents struct {
	offset  int64
	content []byte
}

type FileInfo struct {
	Values []struct {
		Name  string `json:"name"`
		Size  int64  `json:"size"`
		IsDir bool   `json:"dir"`
	} `json:"values"`
}

func (f FileInfo) CreateDirs() {
	for _, d := range f.Values {
		if !d.IsDir {
			continue
		}
		err := os.Mkdir(d.Name, 0750)
		if err != nil && !os.IsExist(err) {
			log.Println(err)
		}
	}
}

func (f FileInfo) DownloadAll(site string) {
	// Open File
	var wg sync.WaitGroup
	rl := make(chan struct{}, 10)
	for _, v := range f.Values {
		if v.IsDir {
			continue
		}
		wg.Add(1)
		rl <- struct{}{}
		go func(file string, size int64) {
			defer wg.Done()
			defer func() { <-rl }()
			downloadFile(site, file, size)
		}(v.Name, v.Size)
	}
	wg.Wait()
}

func downloadFile(site, file string, size int64) {
	f, err := os.Create(file)
	if err != nil {
		log.Println("Creating File ", err)
	}

	site = site + "/" + file
	chunks := size / DownloadSize
	if size%DownloadSize != 0 {
		chunks += 1
	}
	log.Println("Downloading ", site)

	var wg sync.WaitGroup

	in := make(chan FileContents, 100)
	writerComplete := make(chan struct{})
	defer func() { <-writerComplete }()
	go func() {
		defer func() { writerComplete <- struct{}{} }()
		for c := range in {
			_, err := f.WriteAt(c.content, c.offset)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	rl := make(chan struct{}, 50)
	defer close(rl)
	for i := int64(0); i < chunks; i++ {
		wg.Add(1)
		rl <- struct{}{}
		go func(i int64) {
			defer wg.Done()
			defer func() { <-rl }()
			begin := i * DownloadSize
			end := min((i+1)*DownloadSize, size) - 1
			downloadRange(site, begin, end, in)
		}(i)
	}

	wg.Wait()
	close(in)
}

func downloadRange(site string, start, end int64, out chan<- FileContents) {
	req, err := http.NewRequest("GET", site, nil)
	if err != nil {
		log.Println(fmt.Sprintf("creating reqest %s %d-%d", site, start, end), err)
	}
	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	var c http.Client
	resp, err := c.Do(req)
	if err != nil {
		log.Println(fmt.Sprintf("recv reqest %s %d-%d", site, start, end), err)
		return
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	out <- FileContents{start, data}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal(os.Args[0], " http://site filename")
	}

	site := os.Args[1]
	if !strings.HasPrefix(site, "http") {
		site = "http://" + site
	}

	// get files
	resp, err := http.Get(site + "/status/")
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var f FileInfo
	err = json.Unmarshal(data, &f)
	if err != nil {
		log.Println(err)
		return
	}
	f.CreateDirs()
	f.DownloadAll(site + "/download/")
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}

	return b
}
