package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {

	pathS, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	r := regexp.MustCompile(`db_([A-Za-z]+)\.sql`)

	filepath.Walk(pathS, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {

			packageStr := filepath.Base(filepath.Dir(path))

			xs := r.FindStringSubmatch(f.Name())
			if len(xs) > 0 {


				fmt.Println("+",f.Name(), "package", packageStr)
				b,err := ioutil.ReadFile(f.Name())
				if err != nil {
					panic(err)
				}

				s := xs[1]

				f, err := os.Create("sql_" + s + "_generated.go" )
				if err != nil {
					panic(err)
				}
				fmt.Fprintln(f, "package", packageStr)
				fmt.Fprintln(f, "")
				fmt.Fprintf(f, "const SQL%s = `\n", strings.Title(s) )
				f.Write(b)
				fmt.Fprintln(f, "`")
				f.Close()
			}
		}
		return nil
	})





}

