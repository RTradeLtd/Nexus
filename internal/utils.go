package internal

import (
	"fmt"
	"os"
)

func Fatal(msg ...interface{}) {
	fmt.Println(msg...)
	os.Exit(1)
}
