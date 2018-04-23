package main

import (
	"os"
	"strings"
)

const (
	PathSeparator = string(os.PathSeparator)
)

func IsFileExists(name string) bool {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func NormalizeFileName(name string) string {

	name = strings.Replace(name, "/", PathSeparator, len(name))
	name = strings.Replace(name, "\\", PathSeparator, len(name))
	name = strings.TrimSpace(name)
	// name = strings.ToLower(name)

	return name
}

func CreateDirIfNotExist(dir string, perm os.FileMode) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, perm)
		return err
	}
	return nil
}
