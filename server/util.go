package main

import (
	"os"
	"time"
	"strconv"
	"sync"
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

type IdGenerator struct {
	mutex sync.Mutex
	currentNow int64
	currentNowIndex int
}

func NewIdGenerator() *IdGenerator {
	return &IdGenerator{}
}

func (self *IdGenerator) Str(id int64) string {
	return strconv.FormatInt(id, 32)
}

func (self *IdGenerator) Mix(now int64, idx int) int64 {
	return now * (32<<5) + int64(idx)
}

func (self *IdGenerator) GenerateString() string {
	return self.Str(self.Generate())
}

func (self *IdGenerator) Generate() int64 {
	now := time.Now().UnixNano() / 1000000
	nowIndex := 0
	self.mutex.Lock()
	if self.currentNow == now {
		self.currentNowIndex++
	} else {
		self.currentNow = now
		self.currentNowIndex = 0
	}
	nowIndex = self.currentNowIndex
	self.mutex.Unlock()
	return self.Mix(now, nowIndex)
}
