package ankat

import (
	"os"
	"github.com/lxn/win"
	"syscall"
	"path/filepath"
)


func AppDataFileName(filename string) string {
	return filepath.Join(mustAppDataDir(AppName), filename)
}

//func AppFileName(filename string) string {
//	return filepath.Join(mustAppDir(AppName), filename)
//}

const AppName = "ankat"

func mustAppDataDir(app string) string {
	var appDataDir string
	if appDataDir = os.Getenv("MYAPPDATA"); len(appDataDir) == 0 {
		var buf [win.MAX_PATH]uint16
		if !win.SHGetSpecialFolderPath(0, &buf[0], win.CSIDL_APPDATA, false) {
			panic("SHGetSpecialFolderPath failed")
		}
		appDataDir = syscall.UTF16ToString(buf[0:])
	}
	return mustDir(filepath.Join(appDataDir, "Аналитприбор", app))
}

//func mustAppDir(app string) string {
//	usr, err := user.Current()
//	if err != nil {
//		panic(err)
//	}
//	return mustDir(filepath.Join(usr.HomeDir, "."+app))
//}

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