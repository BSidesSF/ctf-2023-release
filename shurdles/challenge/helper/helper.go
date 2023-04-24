package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if path, err := os.Readlink("/proc/self/exe"); err != nil {
		panic(err)
	} else {
		if strings.HasPrefix(path, "/opt") {
			fmt.Println("good shurdles-helper!")
			os.Exit(0)
		} else {
			fmt.Println("this is the wrong shurdles-helper!")
			os.Exit(1)
		}
	}
}
