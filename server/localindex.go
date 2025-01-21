package main

import (
	"fmt"
	"sync"
	"github.com/crumbjp/go-faiss"
	"errors"
	"strings"
	"time"
	pb "github.com/crumbjp/faissdb/server/grpc_replica"
)

const (
	FAISS_TRAINED = "/faiss_trained"
	META_KEY_DB_PREFIX = "DB_"
)

type FaissIndex struct {
	name string
	config Faissconfig
	rwmutex sync.RWMutex
	index faiss.Index
	parameterSpace *faiss.ParameterSpace
}

func newFaissIndex(name string) *FaissIndex {
	faissdb.logger.Info("newFaissIndex(%s)", name)
	faissIndex := &FaissIndex{name: name, config: config.Db.Faiss}
	faissIndex.rwmutex = sync.RWMutex{}
	return faissIndex
}

func (self *FaissIndex) IndexFilePath() string {
	return config.Db.Dbpath + "/" + self.name
}

func (self *FaissIndex) OpenNew() {
	if self.index != nil {
		panic("Already opened")
	}
	faissdb.logger.Info("FaissIndex[%s].OpenNew()", self.name)
	metric := faiss.MetricInnerProduct
	if self.config.Metric == "InnerProduct" {
		metric = faiss.MetricInnerProduct
	} else if self.config.Metric == "L2" {
		metric = faiss.MetricL2
	}
	index, err := faiss.IndexFactory(config.Db.Faiss.Dimension, self.config.Description, metric)
	if err != nil {
		panic(err)
	}
	self.index = index
	self._PostOpen()
}

func (self *FaissIndex) Open(fromTrained bool) error {
	faissdb.logger.Info("FaissIndex[%s].Open()", self.name)
	if self.index != nil {
		panic("Already opened")
	}
	index, err := faiss.ReadIndex(self.IndexFilePath(), faiss.IoFlagMmap)
	if err != nil {
		faissdb.logger.Error("FaissIndex[%s].Open() ReadIndex %v", self.name, err)
	}
	if index == nil {
		if !fromTrained {
			return errors.New(fmt.Sprintf("FaissIndex[%s].Open() Not found", self.name))
		}
		var trainedData []byte
		trainedData, err = ReadFile(TrainedFilePath())
		if err != nil {
			faissdb.logger.Error("FaissIndex[%s].Open() ReadFile %v", self.name, err)
			return err
		}
		err = WriteFile(self.IndexFilePath(), trainedData)
		if err != nil {
			faissdb.logger.Error("FaissIndex[%s].Open() WriteFile %v", self.name, err)
			return err
		}
		index, err = faiss.ReadIndex(self.IndexFilePath(), faiss.IoFlagMmap)
		if err != nil {
			faissdb.logger.Error("FaissIndex[%s].Open() ReadIndex %v", self.name, err)
			return err
		}
	}
	self.index = index
	self._PostOpen()
	return nil
}

func (self *FaissIndex) _PostOpen() {
	var err error
	self.parameterSpace, err = faiss. NewParameterSpace()
	if err != nil {
		panic(err)
	}
	err = self.parameterSpace.SetIndexParameter(self.index, "nprobe", float64(self.config.Nprobe))
	if err != nil {
		panic(err)
	}
	faissdb.metaDB.PutString(META_KEY_DB_PREFIX + self.name, self.name)
	faissdb.logger.Info("FaissIndex[%s]._PostOpen() total: %v", self.name, self.index.Ntotal())
}

func (self *FaissIndex) CloseWithoutLock() {
	faissdb.logger.Info("FaissIndex[%s].CloseWithoutLock()", self.name)
	if self.index != nil {
		self.index.Delete()
		self.index = nil
	}
	if self.parameterSpace != nil {
		self.parameterSpace.Delete()
		self.parameterSpace = nil
	}
}

func (self *FaissIndex) flush(path string) {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	if self.index == nil {
		return
	}
	err := faiss.WriteIndex(self.index, path)
	if err != nil {
		panic(err)
	}
}

func (self *FaissIndex) WriteTrained() {
	faissdb.logger.Info("FaissIndex[%s].WriteTrained() start", self.name)
	self.flush(TrainedFilePath());
	faissdb.logger.Info("FaissIndex[%s].WriteTrained() end", self.name)
}

func (self *FaissIndex) Write() {
	faissdb.logger.Info("FaissIndex[%s].Write() start", self.name)
	self.flush(self.IndexFilePath());
	faissdb.logger.Info("FaissIndex[%s].Write() end", self.name)
}

func (self *FaissIndex) Reset() {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	faissdb.logger.Info("FaissIndex[%s].Reset() start", self.name)
	if self.index != nil {
		self.index.Reset()
	}
	faissdb.logger.Info("FaissIndex[%s].Reset() end", self.name)
}

func (self *FaissIndex) Train(vector []float32) {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	if self.index == nil {
		return
	}
	faissdb.logger.Info("FaissIndex[%s].Train() start", self.name)
	self.index.Reset()
	err := self.index.Train(vector)
	if err != nil {
		panic(err)
	}
	faissdb.logger.Info("FaissIndex[%s].Train() end", self.name)
}

func (self *FaissIndex) AddWithIDs(vectors []float32, xids []int64) error {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	if self.index == nil {
		return nil
	}
	err := self.index.AddWithIDs(vectors, xids)
	if err != nil {
		faissdb.logger.Error("FaissIndex[%s].AddWithIDs() AddWithIDs %v", self.name, err)
	}
	return err
}

func (self *FaissIndex) RemoveIDs(ids []int64) int {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	if self.index == nil {
		return 0
	}
	selector, err := faiss.NewIDSelectorBatch(ids)
	if err != nil {
		panic(err)
	}
	defer selector.Delete()
	var n int
	n, err = self.index.RemoveIDs(selector)
	if err != nil {
		panic(err)
	}
	return n
}

func (self *FaissIndex) Search(vector []float32, n int64) ([]float32, []int64) {
	self.rwmutex.RLock()
	defer self.rwmutex.RUnlock()
	if self.index == nil {
		return []float32{}, []int64{}
	}
	distances, labels, _ := self.index.Search(vector, n)
	return distances, labels
}

func (self *FaissIndex) Ntotal() int64 {
	self.rwmutex.RLock()
	defer self.rwmutex.RUnlock()
	if self.index == nil {
		return 0
	}
	return self.index.Ntotal()
}

var localIndex *LocalIndex

type LocalIndex struct {
	indexes map[string]*FaissIndex
}

func initLocalIndex() {
	faissdb.logger.Info("initLocalIndex()")
	self := &LocalIndex{}
	self.indexes = map[string]*FaissIndex{}
	localIndex = self
}

func (self *LocalIndex) OpenAllIndex() error {
	faissdb.logger.Info("LocalIndex.OpenAllIndex() start")
	defer faissdb.logger.Info("LocalIndex.OpenAllIndex() end")
	it := faissdb.metaDB.db.NewIterator(faissdb.dataDB.defaultReadOptions)
	it.Seek([]byte(META_KEY_DB_PREFIX))
	defer it.Close()
	for it = it; it.Valid(); it.Next() {
		key := it.Key()
		defer key.Free()
		strKey := string(key.Data())
		if !strings.HasPrefix(strKey, META_KEY_DB_PREFIX) {
			break
		}
		value := it.Value()
		defer value.Free()
		collection := string(value.Data())
		self.indexes[collection] = newFaissIndex(collection)
		self.indexes[collection].Open(true)
	}
	return nil
}

func (self *LocalIndex) CloseAll() {
	for _, index := range self.indexes {
		index.rwmutex.Lock()
		index.CloseWithoutLock()
		index.rwmutex.Unlock()
	}
	self.indexes = map[string]*FaissIndex{}
}

func (self *LocalIndex) Ntotal(collection string) int64 {
	if self.indexes[collection] != nil {
		return self.indexes[collection].Ntotal()
	}
	return 0
}

func (self *LocalIndex) Add(faissdbRecord *pb.FaissdbRecord) {
	for _, collection := range faissdbRecord.Collections {
		if self.indexes[collection] == nil {
			self.indexes[collection] = newFaissIndex(collection)
			self.indexes[collection].Open(true)
		}
		err := self.indexes[collection].AddWithIDs(faissdbRecord.V, []int64{faissdbRecord.Id})
		if err != nil {
			panic(err)
		}
	}
}

func (self *LocalIndex) RemoveRaw(collection string, ids []int64) int {
	if self.indexes[collection] != nil {
		self.indexes[collection].RemoveIDs(ids)
	}
	return 0
}

func (self *LocalIndex) Remove(faissdbRecord *pb.FaissdbRecord) int {
	performMain := faissdb.logger.PerformStart("LocalIndex.Remove main")
	faissdb.logger.PerformEnd("LocalIndex.Remove main", performMain)
	for _, collection := range faissdbRecord.Collections {
		if self.indexes[collection] == nil {
			performOpen := faissdb.logger.PerformStart("LocalIndex.Remove Open")
			self.indexes[collection] = newFaissIndex(collection)
			self.indexes[collection].Open(true)
			faissdb.logger.PerformEnd("LocalIndex.Remove Open", performOpen)
		}
		performRemove := faissdb.logger.PerformStart("LocalIndex.Remove Remove")
		self.RemoveRaw(collection, []int64{faissdbRecord.Id})
		faissdb.logger.PerformEnd("LocalIndex.Remove Remove", performRemove)
	}
	return 0
}

func (self *LocalIndex) IsTrained() (bool) {
	return StatFile(TrainedFilePath())
}

func (self *LocalIndex) ResetToTrained() {
	faissdb.logger.Info("LocalIndex.ResetToTrained() start")
	defer faissdb.logger.Info("LocalIndex.ResetToTrained() end")
	data, err := ReadFile(TrainedFilePath())
	if err != nil {
		faissdb.logger.Error("LocalIndex.ResetToTrained() ReadFile(TrainedFilePath()) %v", err)
	}
	for collection, index := range self.indexes {
		index.rwmutex.Lock()
		index.CloseWithoutLock()
		index.rwmutex.Unlock()
		faissdb.logger.Info("LocalIndex.ResetToTrained() Reset index %v", collection)
		err = WriteFile(index.IndexFilePath(), data)
		if err != nil {
			faissdb.logger.Fatal("LocalIndex.ResetToTrained() WriteFile(index.IndexFilePath(), data) %v", err)
		}
		index.Open(false)
	}
}

func (self *LocalIndex) Write() {
	faissdb.logger.Info("LocalIndex.Write() start")
	lastkey := LastKey()
	for _, index := range self.indexes {
		index.Write()
	}
	faissdb.metaDB.PutString("lastkey", lastkey)
	faissdb.logger.Info("LocalIndex.Write() end %s", lastkey)
}

func (self *LocalIndex) SyncFromLocalDb() {
	faissdb.logger.Info("LocalIndex.SyncFromLocalDb() start %s", start)
	defer faissdb.logger.Info("LocalIndex.SyncFromLocalDb() end %s", start)
	it := faissdb.dataDB.db.NewIterator(faissdb.dataDB.defaultReadOptions)
	it.Seek([]byte(""))
	defer it.Close()
	for it = it; it.Valid(); it.Next() {
		key := it.Key()
		defer key.Free()
		value := it.Value()
		defer value.Free()
		faissdbRecord := &pb.FaissdbRecord{}
		DecodeFaissdbRecord(faissdbRecord, value.Data())
		SetRaw(string(key.Data()), faissdbRecord)
	}
	self.Write()
}

func (self *LocalIndex) SyncLocalOplog(start string) {
	faissdb.logger.Info("LocalIndex.SyncLocalOplog() start %s", start)
	defer faissdb.logger.Info("LocalIndex.SyncLocalOplog() end %s", start)
	bulkSize := 10000
	oplog := &Oplog{}
	for ;; {
		keys, values, err := GetCurrentOplog(start, bulkSize)
		if err != nil {
			panic(err)
		}
		for _, value := range values {
			oplog.Decode(value)
			err = ApplyOplog(oplog)
			if err != nil {
				panic(err)
			}
		}
		if len(keys) != bulkSize {
			break
		}
	}
	self.Write()
}

func (self *LocalIndex) Train(trainData []float32) {
	faissdb.logger.Info("LocalIndex.Train() len: %v", len(trainData))
	trainIndex := newFaissIndex("_TRAIN_")
	trainIndex.OpenNew()
	trainIndex.Train(trainData)
	trainIndex.WriteTrained()
	trainIndex.CloseWithoutLock()
}

func (self *LocalIndex) Search(collection string, vector []float32, n int64) ([]float32, []int64) {
	if self.indexes[collection] != nil {
		return self.indexes[collection].Search(vector, n)
	}
	return nil, nil
}

func TrainedFilePath() string {
	return config.Db.Dbpath + FAISS_TRAINED
}

func syncLocalIndexThread() {
	faissdb.logger.Info("syncLocalIndexThread() start")
	for ;; {
		time.Sleep(config.Db.Faiss.Syncinterval * time.Millisecond)
		if faissdb.status == STATUS_READY {
			localIndex.Write()
		}
	}
}

func InitLocalIndex() {
	initLocalIndex()
	localIndex.OpenAllIndex()
	go syncLocalIndexThread()
}

func GapSyncLocalIndex() {
	faissdb.logger.Info("GapSyncLocalIndex() start")
	defer faissdb.logger.Info("GapSyncLocalIndex() end")
	lastkey := LastKey()
	metaLastkey := faissdb.metaDB.GetString("lastkey")
	if lastkey != "" && lastkey != metaLastkey {
		faissdb.logger.Info("GapSyncLocalIndex() Detect gap index(%v) != localdb(%v)", metaLastkey, lastkey)
		localIndex.SyncLocalOplog(metaLastkey)
	}
}
