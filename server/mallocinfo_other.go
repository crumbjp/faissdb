//go:build !linux

package main

import (
	"errors"
)

func MallocInfo() (string, error) {
	return "", errors.New("MallocInfo() not supported on this platform")
}
