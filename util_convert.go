package main

import (
	"strconv"
)

// convert string to int by ignoring any errors
func ForceAtoi(s string) (i int) {
	i64, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		return 0
	}
	return int(i64)
}

// convert string to uint by ignoring any errors
func ForceAtoui(s string) (i uint) {
	i64, err := strconv.ParseUint(s, 10, 0)
	if err != nil {
		return 0
	}
	return uint(i64)
}

// convert string to int by ignoring any errors
func ForceAtoi64(s string) (i int64) {
	i64, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return i64
}

func Itoa(i int) (s string) {
	return strconv.Itoa(i)
}

func I64toa(i int64) (s string) {
	return strconv.FormatInt(i, 10)
}
