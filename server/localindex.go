package main

import (
	"fmt"
	"sync"
	"local.packages/go-faiss" // "github.com/DataIntelligenceCrew/go-faiss"
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
	if(self.name != "main") {
		faissdb.metaDB.PutString(META_KEY_DB_PREFIX + self.name, self.name)
	}
	faissdb.logger.Info("FaissIndex[%s]._PostOpen() total: %v", self.name, self.index.Ntotal())
}

func (self *FaissIndex) Close() {
	faissdb.logger.Info("FaissIndex[%s].Close()", self.name)
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
	if self.index == nil {
		return
	}
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
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
	faissdb.logger.Info("FaissIndex[%s].Reset() start", self.name)
	if self.index != nil {
		self.index.Reset()
	}
	faissdb.logger.Info("FaissIndex[%s].Reset() end", self.name)
}

func (self *FaissIndex) Train(vector []float32) {
	faissdb.logger.Info("FaissIndex[%s].Train() start", self.name)
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	self.index.Reset()
	err := self.index.Train(vector)
	if err != nil {
		panic(err)
	}
	faissdb.logger.Info("FaissIndex[%s].Train() end", self.name)
}

func (self *FaissIndex) AddWithIDs(vectors []float32, xids []int64) error {
	if self.index == nil {
		return nil
	}
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	err := self.index.AddWithIDs(vectors, xids)
	if err != nil {
		faissdb.logger.Error("FaissIndex[%s].AddWithIDs() AddWithIDs %v", self.name, err)
	}
	return err
}

func (self *FaissIndex) RemoveIDs(ids []int64) int {
	if self.index == nil {
		return 0
	}
	selector, err := faiss.NewIDSelectorBatch(ids)
	if err != nil {
		panic(err)
	}
	defer selector.Delete()
	var n int
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	n, err = self.index.RemoveIDs(selector)
	if err != nil {
		panic(err)
	}
	return n
}

func (self *FaissIndex) Search(vector []float32, n int64) ([]float32, []int64) {
	if self.index == nil {
		return []float32{}, []int64{}
	}
	distances, labels, _ := self.index.Search(vector, n)
	return distances, labels
}

func (self *FaissIndex) Ntotal() int64 {
	if self.index == nil {
		return 0
	}
	return self.index.Ntotal()
}

var localIndex *LocalIndex

type LocalIndex struct {
	mainIndex *FaissIndex
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
	self.mainIndex = newFaissIndex("main")
	err := self.mainIndex.Open(false)
	if err != nil {
		faissdb.logger.Error("LocalIndex.OpenAllIndex() Open %v", err)
		self.mainIndex.OpenNew()
	}
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
	self.mainIndex.Close()
	self.mainIndex = nil
	for _, index := range self.indexes {
		index.Close()
	}
	self.indexes = map[string]*FaissIndex{}
}

func (self *LocalIndex) IsTrained() (bool) {
	return self.mainIndex.index.IsTrained()
}

func (self *LocalIndex) Ntotal(collection string) int64 {
	if collection == "" {
		return self.mainIndex.Ntotal()
	}
	if self.indexes[collection] != nil {
		return self.indexes[collection].Ntotal()
	}
	return 0
}

func (self *LocalIndex) Add(faissdbRecord *pb.FaissdbRecord) error {
 	err := self.mainIndex.AddWithIDs(faissdbRecord.V, []int64{faissdbRecord.Id})
	for _, collection := range faissdbRecord.Collections {
		if self.indexes[collection] == nil {
			self.indexes[collection] = newFaissIndex(collection)
			self.indexes[collection].Open(true)
		}
		self.indexes[collection].AddWithIDs(faissdbRecord.V, []int64{faissdbRecord.Id})
	}
	return err
}

func (self *LocalIndex) Remove(faissdbRecord *pb.FaissdbRecord) int {
	performMain := faissdb.logger.PerformStart("LocalIndex.Remove main")
	n := self.mainIndex.RemoveIDs([]int64{faissdbRecord.Id})
	faissdb.logger.PerformEnd("LocalIndex.Remove main", performMain)
	for _, collection := range faissdbRecord.Collections {
		if self.indexes[collection] == nil {
			performOpen := faissdb.logger.PerformStart("LocalIndex.Remove Open")
			self.indexes[collection] = newFaissIndex(collection)
			self.indexes[collection].Open(true)
			faissdb.logger.PerformEnd("LocalIndex.Remove Open", performOpen)
		}
		performRemove := faissdb.logger.PerformStart("LocalIndex.Remove Remove")
		self.indexes[collection].RemoveIDs([]int64{faissdbRecord.Id})
		faissdb.logger.PerformEnd("LocalIndex.Remove Remove", performRemove)
	}
	return n
}

func (self *LocalIndex) ResetToTrained() {
	faissdb.logger.Info("LocalIndex.ResetToTrained() start")
	defer faissdb.logger.Info("LocalIndex.ResetToTrained() end")
	data, err := ReadFile(TrainedFilePath())
	if err != nil {
		faissdb.logger.Error("LocalIndex.ResetToTrained() ReadFile(TrainedFilePath()) %v", err)
	}
 	self.mainIndex.Close()
	err = WriteFile(self.mainIndex.IndexFilePath(), data)
	if err != nil {
		faissdb.logger.Fatal("LocalIndex.ResetToTrained() WriteFile() %v", err)
	}
	self.mainIndex.Open(false)
	for collection, index := range self.indexes {
		index.Close()
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
 	self.mainIndex.Write()
	for _, index := range self.indexes {
		index.Write()
	}
	faissdb.metaDB.PutString("lastkey", lastkey)
	faissdb.logger.Info("LocalIndex.Write() end %s", lastkey)
}

func (self *LocalIndex) SyncFromLocalDb(start string) {
	faissdb.logger.Info("LocalIndex.SyncFromLocalDb() start %s", start)
	defer faissdb.logger.Info("LocalIndex.SyncFromLocalDb() end %s", start)
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
	self.mainIndex.Train(trainData)
	self.mainIndex.WriteTrained()
}

func (self *LocalIndex) Search(collection string, vector []float32, n int64) ([]float32, []int64) {
	if collection == "" {
		return self.mainIndex.Search(vector, n)
	}
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
		localIndex.SyncFromLocalDb(metaLastkey)
	}
}
