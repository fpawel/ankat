package main

import (
	"flag"
)

//go:generate go run ./gen_sql_str/main.go

func main() {
	waitPeer := false
	flag.BoolVar(&waitPeer, "waitpeer", false,  "wait for peer application")
	flag.Parse()
	runApp(waitPeer)
}



