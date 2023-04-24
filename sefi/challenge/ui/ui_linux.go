package main

import (
	"os"
	"path/filepath"
)

func getExeDir() (string, error) {
	if link, err := os.Readlink("/proc/self/exe"); err != nil {
		return "", err
	} else {
		return filepath.Dir(link), nil
	}
}
