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

func (self *LocalDB) Open(dbconfig Dbconfig) {
	self.name = config.Db.Dbpath + self.path
	self.defaultBlockBasedTableOptions = gorocksdb.NewDefaultBlockBasedTableOptions()
	self.defaultBlockBasedTableOptions.SetBlockCache(gorocksdb.NewLRUCache(dbconfig.Capacity))
	self.defaultOptions = gorocksdb.NewDefaultOptions()
	self.defaultOptions.SetBlockBasedTableFactory(self.defaultBlockBasedTableOptions)
	self.defaultOptions.SetCreateIfMissing(true)
	db, err := gorocksdb.OpenDb(self.defaultOptions, self.name)
	if err != nil {
		panic(err)
	}
	self.db = db
	self.defaultReadOptions = gorocksdb.NewDefaultReadOptions()
	self.defaultReadOptions.SetFillCache(false)
	self.defaultWriteOptions = gorocksdb.NewDefaultWriteOptions()
}

func (self *LocalDB) DestroyDb() {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	self.defaultBlockBasedTableOptions.Destroy()
	self.defaultReadOptions.Destroy()
	self.defaultWriteOptions.Destroy()
	self.db.Close()
	gorocksdb.DestroyDb(self.name, self.defaultOptions)
	self.defaultOptions.Destroy()
}

func (self *LocalDB) Close() {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	self.db.Close()
}

func (self *LocalDB) Delete(key string) {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	err := self.db.Delete(self.defaultWriteOptions, []byte(key))
	if err != nil {
		panic(err)
	}
}

func (self *LocalDB) Put(key string, value []byte) {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
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

func GetRawData(startKey string, length int) ([]string, []*gorocksdb.Slice, string) {
	nextKey := ""
	keys := make([]string, length)
	slices := make([]*gorocksdb.Slice, length)
	count := 0
	it := dataDB.db.NewIterator(dataDB.defaultReadOptions)
	it.Seek([]byte(startKey))
	defer it.Close()
	for it = it; it.Valid(); it.Next() {
		key := it.Key()
		value := it.Value()
		defer key.Free()
		if count == length {
			nextKey = string(key.Data())
			break
		}
		keys[count] = string(key.Data())
		slices[count] = value
		count++
	}
	return keys[0:count], slices[0:count], nextKey
}
