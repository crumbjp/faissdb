package main

import (
	"os"
	"time"
	"strconv"
	"strings"
	"fmt"
	"errors"
	"sync"
	"crypto/sha1"
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
	return now * (32 * 32) + int64(idx)
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

func (self *IdGenerator) Parse(id string) (time.Time, int) {
	i64, _ := strconv.ParseInt(id, 32, 64)
	now := i64 / (32 * 32)
	index := i64 % (32 * 32)

	return time.Unix(now / 1000, (now % 1000) * 1000000), int(index)
}

func parseSparseV(v []float32, line string) (error) {
	splited := strings.Split(line, ",")
	for _, str := range splited {
		colonIndex := strings.Index(str, ":")
		if colonIndex < 0 {
			return errors.New("parseSparseV() Invalid sparse vector format")
		}
		key := str[0:colonIndex]
		value := str[colonIndex+1:len(str)]
		i, err := strconv.Atoi(key)
		if err != nil {
			faissdb.logger.Error("parseSparseV() strconv.Atoi(%s) %v", key, err)
			return err
		}
		var f float64
		f, err = strconv.ParseFloat(value, 32)
		if err != nil {
			faissdb.logger.Error("parseSparseV() strconv.ParseFloat(%s, 32) %v", value, err)
			return err
		}
		if i >= config.Db.Faiss.Dimension {
			return errors.New(fmt.Sprintf("parseSparseV() Invalid data dimensions expected: %d actual: %d", config.Db.Faiss.Dimension, i))
		}
		v[i] = float32(f)
	}
	return nil
}

func Sha1(in []byte) string {
	sha1Hash := sha1.New()
	sha1Hash.Write(in)
	return fmt.Sprintf("%x\n", sha1Hash.Sum(nil))
}
