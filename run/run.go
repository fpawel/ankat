package main

import (
	"bytes"
	"github.com/fpawel/goutils/winapp"
	"github.com/lxn/win"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func main() {
	log.SetPrefix("ankat: ")
	log.SetFlags(log.Ltime)
	os.Setenv("GOTRACEBACK", "all")



	exeDir := winapp.MustAppDir(".ankat")

	exe := filepath.Join(exeDir, "ankathost.exe")
	cmd := exec.Command(exe, os.Args[1:]...	)

	printArgs()

	var outB, errB bytes.Buffer
	cmd.Stdout = &outB
	cmd.Stderr = &errB

	if err := cmd.Start(); err != nil {
		log.Panic(err)
	}
	err := cmd.Wait()
	log.Println(outB.String())

	if err != nil {
		strBuff := bytes.NewBuffer(nil)
		dumpCrash(&errB, strBuff)
		log.Println(strBuff.String())

		winapp.MsgBox(strBuff.String(), "АНКАТ", win.MB_ICONERROR)

		errorsDir := winapp.MustDir( filepath.Join(exeDir, "errors") )
		t := time.Now()
		errorFileName := filepath.Join(errorsDir, t.Format("2006_01_02_15_04_05.000")+".log")

		log.Println(errorFileName)

		if err := ioutil.WriteFile(errorFileName, strBuff.Bytes(), 0644); err != nil {
			log.Panic(err)
		}
		log.Println(err)
	}
}





func printArgs(){
	args := make([]interface{}, len(os.Args[1:]))
	for i,v := range os.Args[1:] {
		args[i] = v
	}
	log.Println(args...)
}
