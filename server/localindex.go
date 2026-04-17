package main

import (
	"errors"
	"fmt"
	pb "github.com/crumbjp/faissdb/server/grpc_replica"
	"github.com/crumbjp/go-faiss"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	FAISS_TRAINED      = "/faiss_trained"
	META_KEY_DB_PREFIX = "DB_"
)

type FaissIndex struct {
	name           string
	config         Faissconfig
	rwmutex        sync.RWMutex
	index          faiss.Index
	parameterSpace *faiss.ParameterSpace
	directMapReady bool
}

func newFaissIndex(name string) *FaissIndex {
	faissdb.logger.InfoMem("newFaissIndex(%s)", name)
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
	faissdb.logger.InfoMem("FaissIndex[%s].OpenNew()", self.name)
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
	faissdb.logger.InfoMem("FaissIndex[%s].Open()", self.name)
	if self.index != nil {
		panic("Already opened")
	}
	// Remove any leftover .tmp from a previous crashed flush().
	_ = os.Remove(self.IndexFilePath() + ".tmp")
	fi, statErr := os.Stat(self.IndexFilePath())
	indexFileExists := statErr == nil && fi.Size() > 0
	index, err := faiss.ReadIndex(self.IndexFilePath(), faiss.IoFlagMmap)
	if err != nil {
		faissdb.logger.Error("FaissIndex[%s].Open() ReadIndex %v", self.name, err)
	}
	if index == nil {
		if indexFileExists {
			return errors.New(fmt.Sprintf(
				"FaissIndex[%s].Open() index file exists (size=%d) but ReadIndex failed: corrupted; "+
					"restart with --fullsync to rebuild from local dataDB (primary) "+
					"or remove data directory and bootstrap as secondary: %v",
				self.name, fi.Size(), err))
		}
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
	self.directMapReady = false
	self.parameterSpace, err = faiss.NewParameterSpace()
	if err != nil {
		panic(err)
	}
	err = self.parameterSpace.SetIndexParameter(self.index, "nprobe", float64(self.config.Nprobe))
	if err != nil {
		panic(err)
	}
	if self.config.UseDirectMap() {
		if indexIVF := faiss.AsIVF(self.index); indexIVF != nil {
			err = indexIVF.SetDirectMapType(faiss.DirectMapHashtable)
			if err != nil {
				faissdb.logger.Warn("FaissIndex[%s]._PostOpen() SetDirectMapType() fallback: %v", self.name, err)
			} else {
				self.directMapReady = true
			}
		}
	}
	if self.name != "_TRAIN_" {
		faissdb.metaDB.PutString(META_KEY_DB_PREFIX+self.name, self.name)
	}
	faissdb.logger.InfoMem("FaissIndex[%s]._PostOpen() total: %v", self.name, self.index.Ntotal())
}

func (self *FaissIndex) CloseWithoutLock() {
	faissdb.logger.InfoMem("FaissIndex[%s].CloseWithoutLock()", self.name)
	if self.index != nil {
		self.index.Delete()
		self.index = nil
	}
	if self.parameterSpace != nil {
		self.parameterSpace.Delete()
		self.parameterSpace = nil
	}
	self.directMapReady = false
}

func (self *FaissIndex) flush(path string) {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	if self.index == nil {
		return
	}
	tmp := path + ".tmp"
	if err := faiss.WriteIndexFsync(self.index, tmp); err != nil {
		os.Remove(tmp)
		panic(err)
	}
	if err := os.Rename(tmp, path); err != nil {
		panic(err)
	}
}

func (self *FaissIndex) WriteTrained() {
	faissdb.logger.InfoMem("FaissIndex[%s].WriteTrained() start", self.name)
	self.flush(TrainedFilePath())
	faissdb.logger.InfoMem("FaissIndex[%s].WriteTrained() end", self.name)
}

func (self *FaissIndex) Write() {
	faissdb.logger.InfoMem("FaissIndex[%s].Write() start", self.name)
	self.flush(self.IndexFilePath())
	faissdb.logger.InfoMem("FaissIndex[%s].Write() end", self.name)
}

func (self *FaissIndex) Reset() {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	faissdb.logger.InfoMem("FaissIndex[%s].Reset() start", self.name)
	if self.index != nil {
		self.index.Reset()
	}
	faissdb.logger.InfoMem("FaissIndex[%s].Reset() end", self.name)
}

func (self *FaissIndex) Train(vector []float32) {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	if self.index == nil {
		return
	}
	faissdb.logger.InfoMem("FaissIndex[%s].Train() start", self.name)
	self.index.Reset()
	err := self.index.Train(vector)
	if err != nil {
		panic(err)
	}
	faissdb.logger.InfoMem("FaissIndex[%s].Train() end", self.name)
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
	if len(ids) == 0 {
		return 0
	}
	if self.directMapReady {
		indexIVF := faiss.AsIVF(self.index)
		if indexIVF == nil {
			self.directMapReady = false
		} else {
			n, err := indexIVF.RemoveIDsArray(ids)
			if err != nil {
				faissdb.logger.Warn("FaissIndex[%s].RemoveIDs() RemoveIDsArray() fallback: %v", self.name, err)
				self.directMapReady = false
			} else {
				return n
			}
		}
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

type localIndexMap map[string]*FaissIndex

type LocalIndex struct {
	indexes atomic.Pointer[localIndexMap]
}

func initLocalIndex() {
	faissdb.logger.InfoMem("initLocalIndex()")
	self := &LocalIndex{}
	indexes := localIndexMap{}
	self.indexes.Store(&indexes)
	localIndex = self
}

func (self *LocalIndex) Indexes() localIndexMap {
	indexes := self.indexes.Load()
	if indexes == nil {
		return nil
	}
	return *indexes
}

func (self *LocalIndex) ReplaceIndexes(indexes localIndexMap) {
	self.indexes.Store(&indexes)
}

func (self *LocalIndex) OpenAllIndex() error {
	faissdb.logger.InfoMem("LocalIndex.OpenAllIndex() start")
	defer faissdb.logger.Info("LocalIndex.OpenAllIndex() end")
	indexes := localIndexMap{}
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
		indexes[collection] = newFaissIndex(collection)
		if err := indexes[collection].Open(true); err != nil {
			faissdb.logger.Fatal("LocalIndex.OpenAllIndex() %v", err)
		}
	}
	self.ReplaceIndexes(indexes)
	return nil
}

func (self *LocalIndex) CloseAll() {
	for _, index := range self.Indexes() {
		index.rwmutex.Lock()
		index.CloseWithoutLock()
		index.rwmutex.Unlock()
	}
	self.ReplaceIndexes(localIndexMap{})
}

func (self *LocalIndex) Ntotal(collection string) int64 {
	indexes := self.Indexes()
	if indexes[collection] != nil {
		return indexes[collection].Ntotal()
	}
	return 0
}

func (self *LocalIndex) Add(faissdbRecord *pb.FaissdbRecord) {
	for _, collection := range faissdbRecord.Collections {
		indexes := self.Indexes()
		if indexes[collection] == nil {
			newIndexes := make(localIndexMap, len(indexes)+1)
			for name, index := range indexes {
				newIndexes[name] = index
			}
			newIndexes[collection] = newFaissIndex(collection)
			newIndexes[collection].Open(true)
			self.ReplaceIndexes(newIndexes)
			indexes = newIndexes
		}
		err := indexes[collection].AddWithIDs(faissdbRecord.V, []int64{faissdbRecord.Id})
		if err != nil {
			panic(err)
		}
	}
}

func (self *LocalIndex) RemoveRaw(collection string, ids []int64) int {
	indexes := self.Indexes()
	if indexes[collection] != nil {
		indexes[collection].RemoveIDs(ids)
	}
	return 0
}

func (self *LocalIndex) Remove(faissdbRecord *pb.FaissdbRecord) int {
	performMain := faissdb.logger.PerformStart("LocalIndex.Remove main")
	faissdb.logger.PerformEnd("LocalIndex.Remove main", performMain)
	for _, collection := range faissdbRecord.Collections {
		performRemove := faissdb.logger.PerformStart("LocalIndex.Remove Remove")
		self.RemoveRaw(collection, []int64{faissdbRecord.Id})
		faissdb.logger.PerformEnd("LocalIndex.Remove Remove", performRemove)
	}
	return 0
}

func (self *LocalIndex) IsTrained() bool {
	return StatFile(TrainedFilePath())
}

func (self *LocalIndex) ResetToTrained() {
	faissdb.logger.InfoMem("LocalIndex.ResetToTrained() start")
	defer faissdb.logger.Info("LocalIndex.ResetToTrained() end")
	data, err := ReadFile(TrainedFilePath())
	if err != nil {
		faissdb.logger.Error("LocalIndex.ResetToTrained() ReadFile(TrainedFilePath()) %v", err)
	}
	for collection, index := range self.Indexes() {
		index.rwmutex.Lock()
		index.CloseWithoutLock()
		index.rwmutex.Unlock()
		faissdb.logger.InfoMem("LocalIndex.ResetToTrained() Reset index %v", collection)
		err = WriteFile(index.IndexFilePath(), data)
		if err != nil {
			faissdb.logger.Fatal("LocalIndex.ResetToTrained() WriteFile(index.IndexFilePath(), data) %v", err)
		}
		index.Open(false)
	}
}

func (self *LocalIndex) Write() {
	faissdb.logger.InfoMem("LocalIndex.Write() start")
	lastkey := LastKey()
	for _, index := range self.Indexes() {
		index.Write()
	}
	faissdb.metaDB.PutString("lastkey", lastkey)
	faissdb.logger.InfoMem("LocalIndex.Write() end %s", lastkey)
}

func (self *LocalIndex) SyncFromLocalDb() {
	faissdb.logger.InfoMem("LocalIndex.SyncFromLocalDb() start")
	defer faissdb.logger.Info("LocalIndex.SyncFromLocalDb() end")
	it := faissdb.dataDB.db.NewIterator(faissdb.dataDB.defaultReadOptions)
	it.Seek([]byte(""))
	defer it.Close()
	count := 0
	for it = it; it.Valid(); it.Next() {
		key := it.Key()
		value := it.Value()
		faissdbRecord := &pb.FaissdbRecord{}
		DecodeFaissdbRecord(faissdbRecord, value.Data())
		SyncRaw(string(key.Data()), faissdbRecord)
		key.Free()
		value.Free()
		count++
		if count % 10000 == 0 {
			faissdb.logger.InfoMem("LocalIndex.SyncFromLocalDb() synced %d", count)
		}
	}
	faissdb.logger.InfoMem("LocalIndex.SyncFromLocalDb() sync complete: %d records", count)
	self.Write()
}

func (self *LocalIndex) SyncLocalOplog(start string) {
	faissdb.logger.InfoMem("LocalIndex.SyncLocalOplog() start %s", start)
	defer faissdb.logger.Info("LocalIndex.SyncLocalOplog() end %s", start)
	bulkSize := 10000
	oplog := &Oplog{}
	needFullSync := false
	for {
		keys, values, err := GetCurrentOplog(start, bulkSize)
		if err != nil {
			panic(err)
		}
		for _, value := range values {
			oplog.Decode(value)
			err = ApplyOplog(oplog)
			if err == ErrNeedFullSync {
				faissdb.logger.Info("LocalIndex.SyncLocalOplog() OP_FULLSYNC detected, need FullLocalSync")
				needFullSync = true
				break
			}
			if err != nil {
				panic(err)
			}
		}
		if needFullSync || len(keys) != bulkSize {
			break
		}
	}
	if needFullSync {
		FullLocalSync()
	} else {
		self.Write()
	}
}

func (self *LocalIndex) Train(trainData []float32) {
	faissdb.logger.InfoMem("LocalIndex.Train() len: %v", len(trainData))
	trainIndex := newFaissIndex("_TRAIN_")
	trainIndex.OpenNew()
	trainIndex.Train(trainData)
	trainIndex.WriteTrained()
	trainIndex.CloseWithoutLock()
}

func (self *LocalIndex) Search(collection string, vector []float32, n int64) ([]float32, []int64) {
	indexes := self.Indexes()
	if indexes[collection] != nil {
		return indexes[collection].Search(vector, n)
	}
	return nil, nil
}

func TrainedFilePath() string {
	return config.Db.Dbpath + FAISS_TRAINED
}

func syncLocalIndexThread() {
	faissdb.logger.InfoMem("syncLocalIndexThread() start")
	for {
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
	faissdb.logger.InfoMem("GapSyncLocalIndex() start")
	defer faissdb.logger.Info("GapSyncLocalIndex() end")
	lastkey := LastKey()
	metaLastkey := faissdb.metaDB.GetString("lastkey")
	if lastkey != "" && lastkey != metaLastkey {
		faissdb.logger.InfoMem("GapSyncLocalIndex() Detect gap index(%v) != localdb(%v)", metaLastkey, lastkey)
		localIndex.SyncLocalOplog(metaLastkey)
	}
}
