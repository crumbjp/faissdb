package main

import (
	"fmt"
	"time"
	"errors"
	"bytes"
	"io/ioutil"
	"encoding/binary"
)

type Oplog struct {
	op int8
	key string
	d []byte
}

const (
	OP_SYSTEM = int8(0)
	OP_SET = int8(1)
	OP_DEL = int8(2)
	OP_DROPALL = int8(3)
)

func (self *Oplog) Encode() ([]byte, error) {
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.LittleEndian, self.op)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.LittleEndian, int16(len(self.key)))
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.LittleEndian, []byte(self.key))
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.LittleEndian, self.d)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (self *Oplog) Decode(b []byte) error {
	buffer := bytes.NewReader(b)
	err := binary.Read(buffer, binary.LittleEndian, &self.op)
	if err != nil {
		return err
	}
	var length int16
	err = binary.Read(buffer, binary.LittleEndian, &length)
	if err != nil {
		return err
	}
	keyBuffer := make([]byte, int(length))
	err = binary.Read(buffer, binary.LittleEndian, keyBuffer)
	if err != nil {
		return err
	}
	self.key = string(keyBuffer)
	self.d, err = ioutil.ReadAll(buffer)
	return nil
}

func deleteOpLogThread() {
	for ;; {
		time.Sleep(10000 * time.Millisecond)
		if faissdb.status == STATUS_READY {
			continue
		}
		deleteMs := (time.Now().UnixNano() / 1000000) - (int64(config.Oplog.Term) * 1000)
		lastKey :=	LastKey()
		deleteLastKey := faissdb.oplogKeyGenerator.Str(faissdb.oplogKeyGenerator.Mix(deleteMs, 0))
		it := faissdb.oplogDB.db.NewIterator(faissdb.dataDB.defaultReadOptions)
		it.Seek([]byte(""))
		defer it.Close()
		for it = it; it.Valid(); it.Next() {
			key := it.Key()
			defer key.Free()
			oplogKey := string(key.Data())
			if oplogKey == lastKey {
				break // Keep at least 1 log
			}
			if deleteLastKey <= oplogKey {
				break
			}
			faissdb.oplogDB.Delete(oplogKey)
		}
		if IsPrimary() {
			PutOplog(OP_SYSTEM, "", []byte("deleteOpLogThread"))
		}
	}
}

func LastKey() string {
	it := faissdb.oplogDB.db.NewIterator(faissdb.dataDB.defaultReadOptions)
	it.SeekToLast()
	lastKey := string(it.Key().Data())
	defer it.Close()
	return lastKey
}

func GetCurrentOplog(startLogkey string, length int) ([]string, [][]byte, error){
	logkeys := make([]string, length)
	values := make([][]byte, length)
	first := true
	count := 0
	it := faissdb.oplogDB.db.NewIterator(faissdb.oplogDB.defaultReadOptions)
	it.Seek([]byte(startLogkey))
	defer it.Close()
	for it = it; it.Valid(); it.Next() {
		key := it.Key()
		value := it.Value()
		defer key.Free()
		defer value.Free()
		strLogkey := string(key.Data())
		if first {
			if startLogkey != "" && startLogkey != strLogkey {
				return nil, nil, errors.New(fmt.Sprintf("GetCurrentOplog() Stale oplog expected: %s  actual: %s", startLogkey, strLogkey))
			}
			first = false
			continue
		}
		logkeys[count] = strLogkey
		data := value.Data()
		values[count] = make([]byte, len(data))
		copy(values[count], data)
		count++
		if count == length {
			break
		}
	}
	return logkeys[0:count], values[0:count], nil
}

func InitOplog() {
	faissdb.oplogDB = newLocalDB("/log")
	faissdb.oplogDB.Open(&config.Db.Oplogdb)
	faissdb.oplogKeyGenerator = NewIdGenerator()
	go deleteOpLogThread()
}

func PutOplogWithKey(logKey string, op int8, key string, d []byte) {
	oplog := Oplog{op: op, key: key, d: d}
	encodedOplog, _ := oplog.Encode()
	faissdb.oplogDB.Put(logKey, encodedOplog)
}

func PutOplog(op int8, key string, d []byte) {
	logKey := faissdb.oplogKeyGenerator.GenerateString()
	PutOplogWithKey(logKey, op, key, d)
}

func ReadFaissTrained() ([]byte, error) {
	if faissdb.status == STATUS_TRAINING {
		return nil, errors.New("ReadFaissTrained() Training now")
	}
	data, err := ReadFile(TrainedFilePath())
	if err != nil {
    faissdb.logger.Error("ReadFaissTrained() ReadFile(TrainedFilePath()) %v", err)
	}
	return data, err
}
