package main

import (
	"os"
)

func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	data := make([]byte, stat.Size())
	_, readErr := file.Read(data)
	if readErr != nil {
		return nil, readErr
	}
	return data, nil
}

func WriteFile(path string, data []byte) (error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, writeErr := file.Write(data)
	if writeErr != nil {
		return writeErr
	}
	return nil
}
