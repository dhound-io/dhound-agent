// +build windows
package main

import (
	"fmt"
	"os"
	"reflect"
	"syscall"
)

func Executable() (string, error) {
	return os.Executable()
}

func GetFileOsUniqueKey(path string) string {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return ""
	}

	// loading samefile function calls loadingFileId() on info.Sys().(*os.fileStat)
	os.SameFile(info, info)

	fileStat := reflect.ValueOf(info).Elem()

	// the following fields unique identifies in file system despite the file name
	idxhi := fileStat.FieldByName("idxhi").Uint()
	idxlo := fileStat.FieldByName("idxlo").Uint()
	vol := fileStat.FieldByName("vol").Uint()

	key := fmt.Sprintf("%d_%d_%d", idxhi, idxlo, vol)

	return key
}

func ReadOpen(path string) (*os.File, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("File '%s' not found. Error: %v", path, syscall.ERROR_FILE_NOT_FOUND)
	}

	pathp, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, fmt.Errorf("Error converting to UTF16: %v", err)
	}

	var access uint32
	access = syscall.GENERIC_READ

	sharemode := uint32(syscall.FILE_SHARE_READ | syscall.FILE_SHARE_WRITE | syscall.FILE_SHARE_DELETE)

	var sa *syscall.SecurityAttributes

	var createmode uint32

	createmode = syscall.OPEN_EXISTING

	handle, err := syscall.CreateFile(pathp, access, sharemode, sa, createmode, syscall.FILE_ATTRIBUTE_NORMAL, 0)

	if err != nil {
		return nil, fmt.Errorf("Error creating file '%s': %v", pathp, err)
	}

	return os.NewFile(uintptr(handle), path), nil
}
