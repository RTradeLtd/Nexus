package main

import (
	"fmt"
	"os"
)

func fatal(msg ...interface{}) {
	fmt.Println(msg...)
	os.Exit(1)
}
