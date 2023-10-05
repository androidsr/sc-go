package sctail

import (
	"fmt"
	"testing"
)

func TestXxx(t *testing.T) {
	m := New("/mnt/d/test.txt")
	m.Start(func(line string) {
		fmt.Print(line)
	})
	fmt.Println("结束")
}
