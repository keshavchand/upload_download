package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type FileInfo struct {
	name string
	size int64
  dir  bool
}

func (f FileInfo) ToJson() string {
  dir := "false"
  if f.dir {
    dir = "true"
  }
  return fmt.Sprintf(`{ "name": "%s", "size": %d, "dir": %s}`, f.name, f.size, dir)
}

func scanDir(s string) []FileInfo {
	var final []FileInfo
	entries, err := os.ReadDir(s)
	if err != nil {
		return final
	}

	for _, e := range entries {
		name := s + "/" + e.Name()
		if e.IsDir() {
      final = append(final, FileInfo{name, -1, true})
			final = append(final, scanDir(name)...)
			continue
		}

		info, err := e.Info()
		if err != nil {
			log.Println(err)
			continue
		}

		final = append(final, FileInfo{name, info.Size(), false})
	}

	return final
}

func WriteFull(w io.Writer, p []byte) error {
	for len(p) > 0 {
		n, err := w.Write(p)
		if err != nil {
			return err
		}
		p = p[n:]
	}
	return nil
}

func main() {
	log.Println("Serving")
	http.HandleFunc("/status/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		ent := scanDir(".")

		io.WriteString(w, `{ "values" : [`)
		defer io.WriteString(w, `]}`)
		for i, e := range ent {
			if i > 0 {
				w.Write([]byte(","))
			}
			io.WriteString(w, e.ToJson())
		}
	})

	http.HandleFunc("/download/", func(w http.ResponseWriter, r *http.Request) {
		loc := strings.TrimPrefix(r.URL.Path, "/download/")
		http.ServeFile(w, r, loc)
	})
	log.Fatal(http.ListenAndServe(":8000", nil))
}
