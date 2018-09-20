package main

import (
	"github.com/fpawel/goutils/panichook"
	"github.com/fpawel/goutils/winapp"
	"github.com/lxn/win"
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
	r, err := panichook.Run(exeFileName, "-waitpeer=true")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(":\n" + r.Stdout.String())
	if r.Error == nil {
		return
	}
	log.Println("ERROR:", r.Error, ":\n", r.Panic.String())
	log.Println("ERROR LOG FILE:", r.ExeFileName)
	str := r.Panic.String() + "\n" + r.Error.Error() + "\n" + r.ErrorFileName
	winapp.MsgBox(str, "АНКАТ", win.MB_ICONERROR)
}
