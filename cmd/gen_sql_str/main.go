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

	pathS = filepath.Join(pathS, "..", "..")

	r := regexp.MustCompile(`db_([A-Za-z]+)\.sql`)

	filepath.Walk(pathS, func(path string, f os.FileInfo, _ error) error {
		if f == nil || f.IsDir() {
			return nil
		}
		xs := r.FindStringSubmatch(f.Name())
		if len(xs)==  0 {
			return nil
		}
		b,err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}

		s := xs[1]

		fileName := "sql_" + s + "_generated.go"

		fmt.Println("+",f.Name(), ":",)
		fmt.Println("\t", fileName)

		fs, err := os.Create(filepath.Join(filepath.Dir(path), fileName) )
		if err != nil {
			panic(err)
		}
		defer fs.Close()

		fmt.Fprintln(fs, "package", filepath.Base(filepath.Dir(path)))
		fmt.Fprintln(fs, "")
		fmt.Fprintf(fs, "const SQL%s = `\n", strings.Title(s) )
		fs.Write(b)
		fmt.Fprintln(fs, "`")

		return nil
	})





}

