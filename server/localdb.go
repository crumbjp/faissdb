package main

import (
	"sync"
	"github.com/tecbot/gorocksdb"
	"encoding/binary"
	"bytes"
)

type LocalDB struct {
	rwmutex sync.RWMutex
	path string
	name string
	defaultBlockBasedTableOptions *gorocksdb.BlockBasedTableOptions
	defaultOptions *gorocksdb.Options
	db *gorocksdb.DB
	defaultReadOptions *gorocksdb.ReadOptions
	defaultWriteOptions *gorocksdb.WriteOptions
}

func newLocalDB(path string) *LocalDB {
	localDb := &LocalDB{}
	localDb.rwmutex = sync.RWMutex{}
	localDb.path = path
	return localDb
}

func (self *LocalDB) Open(dbconfig *Dbconfig) {
	faissdb.logger.Info("LocalDB[%s].Open() start", self.name)
	defer faissdb.logger.Info("LocalDB[%s].Open() end", self.name)
	self.name = config.Db.Dbpath + self.path
	self.defaultBlockBasedTableOptions = gorocksdb.NewDefaultBlockBasedTableOptions()
	self.defaultBlockBasedTableOptions.SetBlockCache(gorocksdb.NewLRUCache(dbconfig.Capacity))
	self.defaultOptions = gorocksdb.NewDefaultOptions()
	self.defaultOptions.SetBlockBasedTableFactory(self.defaultBlockBasedTableOptions)
	self.defaultOptions.SetCreateIfMissing(true)
	db, err := gorocksdb.OpenDb(self.defaultOptions, self.name)
	if err != nil {
		faissdb.logger.Fatal("LocalDB[%s].Open() gorocksdb.OpenDb() %v", self.name, err)
	}
	self.db = db
	self.defaultReadOptions = gorocksdb.NewDefaultReadOptions()
	self.defaultReadOptions.SetFillCache(false)
	self.defaultWriteOptions = gorocksdb.NewDefaultWriteOptions()
}

func (self *LocalDB) DestroyDb() {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	faissdb.logger.Info("LocalDB[%s].DestroyDb() start", self.name)
	defer faissdb.logger.Info("LocalDB[%s].DestroyDb() end", self.name)
	self.defaultBlockBasedTableOptions.Destroy()
	self.defaultReadOptions.Destroy()
	self.defaultWriteOptions.Destroy()
	self.db.Close()
	gorocksdb.DestroyDb(self.name, self.defaultOptions)
	self.defaultOptions.Destroy()
	self.defaultBlockBasedTableOptions = nil
	self.defaultReadOptions = nil
	self.defaultWriteOptions = nil
	self.defaultOptions = nil
	self.db = nil
}

func (self *LocalDB) Close() {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	faissdb.logger.Info("LocalDB[%s].Close() start", self.name)
	defer faissdb.logger.Info("LocalDB[%s].Close() end", self.name)
	self.defaultBlockBasedTableOptions.Destroy()
	self.defaultReadOptions.Destroy()
	self.defaultWriteOptions.Destroy()
	self.db.Close()
	self.defaultOptions.Destroy()
	self.defaultBlockBasedTableOptions = nil
	self.defaultReadOptions = nil
	self.defaultWriteOptions = nil
	self.defaultOptions = nil
	self.db = nil
}

func (self *LocalDB) Delete(key string) {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	if self.db == nil {
		return
	}
	err := self.db.Delete(self.defaultWriteOptions, []byte(key))
	if err != nil {
		panic(err)
	}
}

func (self *LocalDB) Put(key string, value []byte) {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	if self.db == nil {
		return
	}
	err := self.db.Put(self.defaultWriteOptions, []byte(key), value)
	if err != nil {
		panic(err)
	}
}

func (self *LocalDB) PutString(key string, value string) {
	self.Put(key, []byte(value))
}

func (self *LocalDB) PutInt64(key string, value int64) {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.LittleEndian, &value)
	self.Put(key, buffer.Bytes())
}

func (self *LocalDB) Get(key string) *gorocksdb.Slice {
	self.rwmutex.RLock()
	defer self.rwmutex.RUnlock()
	if self.db == nil {
		return nil
	}
	value, err := self.db.Get(self.defaultReadOptions, []byte(key))
	if err != nil {
		panic(err)
	}
	return value
}

func (self *LocalDB) GetString(key string) string {
	value := self.Get(key)
	defer value.Free()
	return string(value.Data())
}

func (self *LocalDB) GetInt64(key string) *int64 {
	var result int64
	value := self.Get(key)
	defer value.Free()
	buffer := bytes.NewReader(value.Data())
	err := binary.Read(buffer, binary.LittleEndian, &result)
	if err != nil {
		return nil
	}
	return &result
}

func (self *LocalDB) GetRawData(startKey string, length int) ([]string, [][]byte, string) {
	self.rwmutex.RLock()
	defer self.rwmutex.RUnlock()
	if self.db == nil {
		return nil, nil, ""
	}
	nextKey := ""
	keys := make([]string, length)
	values := make([][]byte, length)
	count := 0
	it := self.db.NewIterator(self.defaultReadOptions)
	it.Seek([]byte(startKey))
	defer it.Close()
	for it = it; it.Valid(); it.Next() {
		key := it.Key()
		defer key.Free()
		if count == length {
			nextKey = string(key.Data())
			break
		}
		value := it.Value()
		defer value.Free()
		keys[count] = string(key.Data())
		data := value.Data()
		values[count] = make([]byte, len(data))
		copy(values[count], data)
		count++
	}
	return keys[0:count], values[0:count], nextKey
}
