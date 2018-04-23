// +build !windows

package main

import (
	"fmt"
	"os"
	"syscall"
)

func Executable() (string, error) {
	return "not supported", nil
}

func GetFileOsUniqueKey(path string) string {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return ""
	}

	stat := info.Sys().(*syscall.Stat_t)

	// the following fields unique identifies in file system despite the file name
	inode := uint64(stat.Ino)
	device := uint64(stat.Dev)

	key := fmt.Sprintf("%d_%d", inode, device)

	return key
}

func ReadOpen(path string) (*os.File, error) {
	flag := os.O_RDONLY
	perm := os.FileMode(0)
	return os.OpenFile(path, flag, perm)
}
