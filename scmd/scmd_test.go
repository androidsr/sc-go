package scmd

import (
	"fmt"
	"testing"
)

func Test_New(t *testing.T) {
	c := New(func(shell, output string) bool {
		fmt.Println(output)
		return true
	})
	c.Command("cd /mnt/d")
	c.Command("./test.sh")
}
