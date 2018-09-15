package uiworks

import (
	"os"
	"encoding/json"
)

func readJsonFromFile(fileName string, v interface{}) error {
	f, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	err = json.NewDecoder( f ).Decode(v)
	if err != nil {
		panic(err)
	}
	return f.Close()
}

func writeJsonToFile(fileName string, v interface{}) error {
	f, err := os.OpenFile(fileName, os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	if err = enc.Encode(v) ; err != nil {
		panic(err)
	}
	return f.Close()
}
