package main

import (
	"strconv"
	"fmt"
	"time"
	"bytes"
	"encoding/binary"
)

type Oplog struct {
	op int8
	d []byte
}

const (
	OP_SET = int8(1)
)

func (self *Oplog) Encode() ([]byte, error) {
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.LittleEndian, self.op)
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
	self.d = b[1:len(b)]
	return nil
}

var oplogDB *LocalDB

func deleteOpLogThread() {
	for ;; {
		deleteMs := (time.Now().UnixNano() / 1000000) - (int64(config.Oplog.Term) * 1000)
		lastKey := strLogKey(deleteMs, 0)
		it := oplogDB.db.NewIterator(dataDB.defaultReadOptions)
		it.Seek([]byte(""))
		defer it.Close()
		for it = it; it.Valid(); it.Next() {
			key := it.Key()
			defer key.Free()
			oplogKey := string(key.Data())
			if lastKey <= oplogKey {
				break
			}
			oplogDB.Delete(oplogKey)
		}
		time.Sleep(10000 * time.Millisecond)
	}
}

func intOplog() {
	oplogDB = newLocalDB("/log")
	oplogDB.Open(config.Db.Oplogdb)
	go deleteOpLogThread()
}

func strLogKey(t int64, i int) string{
	return fmt.Sprintf("%09s%02s", strconv.FormatInt(t, 32), strconv.FormatInt(int64(i), 32))
}

func generateLogKey() string{
	now := time.Now().UnixNano() / 1000000
	nowIndex := 0
	nowMutex.Lock()
	if currentNow == now {
		currentNowIndex++
	} else {
		currentNow = now
		currentNowIndex = 0
	}
	nowIndex = currentNowIndex
	nowMutex.Unlock()
	return strLogKey(now, nowIndex)
}
