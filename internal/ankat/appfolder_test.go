package ankat

import (
	"fmt"
	"testing"
)

func Test_mustAppDataDir(t *testing.T) {
	fmt.Println(MustAppDir())
	fmt.Println(MustAppDataDir())
}
