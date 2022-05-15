package main

import (
	"errors"
	"log"
	"net/http"
	"os"
)

func checkIfExists(filename string) error {
	fileinfo, err := os.Stat(filename)
	if err != nil {
		return err
	}
	if fileinfo.IsDir() {
		return errors.New("Not a file")
	}

	return nil
}
func main() {
	if len(os.Args) < 2 {
		log.Fatal(os.Args[0])
	}
	filename := os.Args[1]
	fileErr := checkIfExists(filename)
	if fileErr != nil {
		log.Fatal(fileErr)
	}

  log.Println("Serving")
	http.HandleFunc("/", ServeFile(filename))
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func ServeFile(filename string) func(http.ResponseWriter, *http.Request){
  return func (w http.ResponseWriter, r *http.Request) {
    log.Println("Connection")
    http.ServeFile(w, r, filename)
  }
}
