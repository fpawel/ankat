package main

import (
	"flag"
)

//go:generate go run ./gen_sql_str/main.go

func main() {
	waitPeer := false
	flag.Parse()
	flag.BoolVar(&waitPeer, "wait_peer", false,  "wait for peer application")
	runApp(waitPeer)
}

