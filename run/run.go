package main

import (
	"github.com/fpawel/goutils/panichook"
	"log"
	"os"
	"path/filepath"
)

func main() {
	log.SetPrefix("RUN ANKAT: ")
	log.SetFlags(log.Ltime)
	//os.Setenv("GOTRACEBACK", "all")
	exeDir := filepath.Dir(os.Args[0])
	exeFileName := filepath.Join(exeDir, "ankathost.exe")
	panichook.Run(exeFileName, )
}
