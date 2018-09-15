package ankat

import (
	"github.com/lxn/win"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
)


func AppDataFileName(filename string) string {
	return filepath.Join(MustAppDataDir(), filename)
}

func AppFileName(filename string) string {
	return filepath.Join(MustAppDir(), filename)
}

const AppName = "ankat"

func MustAppDataDir() string {
	var appDataDir string
	if appDataDir = os.Getenv("MYAPPDATA"); len(appDataDir) == 0 {
		var buf [win.MAX_PATH]uint16
		if !win.SHGetSpecialFolderPath(0, &buf[0], win.CSIDL_APPDATA, false) {
			panic("SHGetSpecialFolderPath failed")
		}
		appDataDir = syscall.UTF16ToString(buf[0:])
	}
	return mustDir(filepath.Join(appDataDir, "Аналитприбор", AppName))
}

func MustAppDir() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return mustDir(filepath.Join(usr.HomeDir, "."+AppName))
}

func mustDir(dir string) string {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) { // создать каталог если его нет
			os.Mkdir(dir, os.ModePerm)
		} else {
			panic(err)
		}
	}
	return dir
}