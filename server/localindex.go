package main

import (
	"sync"
	"local.packages/go-faiss" // "github.com/DataIntelligenceCrew/go-faiss"
	"log"
	"strconv"
	"time"
	pb "github.com/crumbjp/faissdb/server/grpc_replica"
)

const (
	FAISS_TRAINED = "/faiss_trained"
)

type FaissIndex struct {
	name string
	config Faissconfig
	rwmutex sync.RWMutex
	index faiss.Index
	parameterSpace *faiss.ParameterSpace
}

func newFaissIndex(name string) *FaissIndex {
	return &FaissIndex{name: name, config: config.Db.Faiss}
}

func (self *FaissIndex) IndexFilePath() string {
	return config.Db.Dbpath + "/" + self.name
}

func (self *FaissIndex) Open() {
	log.Printf("FaissIndex.Open() %v", self.name)
	self.rwmutex = sync.RWMutex{}
	metric := faiss.MetricInnerProduct
	if self.config.Metric == "InnerProduct" {
		metric = faiss.MetricInnerProduct
	} else if self.config.Metric == "L2" {
		metric = faiss.MetricL2
	} else if self.config.Metric == "L1" {
		metric = faiss.MetricL1
	} else if self.config.Metric == "Linf" {
		metric = faiss.MetricLinf
	} else if self.config.Metric == "Lp" {
		metric = faiss.MetricLp
	} else if self.config.Metric == "Canberra" {
		metric = faiss.MetricCanberra
	} else if self.config.Metric == "BrayCurtis" {
		metric = faiss.MetricBrayCurtis
	} else if self.config.Metric == "JensenShannon" {
		metric = faiss.MetricJensenShannon
	}
	index, err := faiss.ReadIndex(self.IndexFilePath(), faiss.IoFlagMmap)
	if err != nil {
		log.Println(err)
	}
	if index == nil {
		index, err = faiss.IndexFactory(config.Db.Faiss.Dimension, self.config.Description, metric)
		if err != nil {
			panic(err)
		}
	}
	self.index = index
	log.Println("ReadIndex total: ", self.index.Ntotal())
	self.parameterSpace, err = faiss. NewParameterSpace()
	if err != nil {
		log.Println(err)
	}
	err = self.parameterSpace.SetIndexParameter(self.index, "nprobe", float64(self.config.Nprobe))
	if err != nil {
		log.Println(err)
	}
}

func (self *FaissIndex) Close() {
	log.Printf("FaissIndex.Close() %v", self.name)
	if self.index != nil {
		self.index.Delete()
	}
	if self.parameterSpace != nil {
		self.parameterSpace.Delete()
	}
}

func (self *FaissIndex) flush(path string) {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	err := faiss.WriteIndex(self.index, path)
	if err != nil {
		panic(err)
	}
}

func (self *FaissIndex) WriteTrained() {
	log.Printf("FaissIndex.WriteTrained() %v start")
	self.flush(TrainedFilePath());
	log.Printf("FaissIndex.WriteTrained() %v end")
}

func (self *FaissIndex) Write() {
	log.Printf("FaissIndex.Write() %v start", self.name)
	lastkey := LastKey()
	self.flush(self.IndexFilePath());
	metaDB.PutString("lastkey", lastkey)
	log.Printf("FaissIndex.Write() %v end", self.name)
}

func (self *FaissIndex) Reset() {
	if self.index != nil {
		self.index.Reset()
	}
}

func (self *FaissIndex) Train(vector []float32) {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	self.index.Reset()
	err := self.index.Train(vector)
	if err != nil {
		panic(err)
	}
}

func (self *FaissIndex) AddWithIDs(vectors []float32, xids []int64) error {
	if self.index == nil {
		return nil
	}
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	err := self.index.AddWithIDs(vectors, xids)
	if err != nil {
		log.Printf("AddWithIDs() %v %v", self.name, err)
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

func newLocalIndex() *LocalIndex {
	self := &LocalIndex{}
	self.mainIndex = newFaissIndex("main")
	self.indexes = map[string]*FaissIndex{}
	for _, collection := range config.Db.Faiss.Collections {
		self.indexes[collection] = newFaissIndex(collection)
	}
	return self
}

func (self *LocalIndex) IsTrained() (bool) {
	return self.mainIndex.index.IsTrained()
}

func (self *LocalIndex) Ntotal() int64 {
	return self.mainIndex.Ntotal()
}

func (self *LocalIndex) Open() {
	self.mainIndex.Open()
	log.Printf("Open mainIndex IsTraind: %v", self.mainIndex.index.IsTrained())
	if self.mainIndex.index.IsTrained() {
		for _, collection := range config.Db.Faiss.Collections {
			if self.indexes[collection] != nil {
				self.indexes[collection].Open()
			}
		}
	}
}

func (self *LocalIndex) Add(faissdbRecord *pb.FaissdbRecord) error {
 	err := self.mainIndex.AddWithIDs(faissdbRecord.V, []int64{faissdbRecord.Id})
	for _, collection := range faissdbRecord.Collections {
		if self.indexes[collection] != nil {
			self.indexes[collection].AddWithIDs(faissdbRecord.V, []int64{faissdbRecord.Id})
		}
	}
	return err
}

func (self *LocalIndex) Remove(id int64) int {
 	n := self.mainIndex.RemoveIDs([]int64{id})
	for _, collection := range config.Db.Faiss.Collections {
		if self.indexes[collection] != nil {
			self.indexes[collection].RemoveIDs([]int64{id})
		}
	}
	return n
}

func (self *LocalIndex) ResetToTrained() {
	data, err := ReadFile(TrainedFilePath())
	if err != nil {
		log.Printf("Trained index file not found", err)
	}
 	self.mainIndex.Close()
	err = WriteFile(self.mainIndex.IndexFilePath(), data)
	if err != nil {
		log.Fatalf("WriteFile() %v", err)
	}
 	self.mainIndex.Open()
	for _, collection := range config.Db.Faiss.Collections {
		if self.indexes[collection] != nil {
			log.Printf("Close %s", collection)
			self.indexes[collection].Close()
		}
		log.Printf("Open by traind file %v", collection)
		err = WriteFile(self.indexes[collection].IndexFilePath(), data)
		if err != nil {
			log.Fatalf("WriteFile() %v", err)
		}
		self.indexes[collection].Open()
	}
}

func (self *LocalIndex) Write() {
 	self.mainIndex.Write()
	for _, collection := range config.Db.Faiss.Collections {
		if self.indexes[collection] != nil {
			self.indexes[collection].Write()
		}
	}
}

func (self *LocalIndex) SyncFromLocalDb(start string) {
	log.Println("SyncFromLocalDb()")
	tmpFaissdbRecord := &pb.FaissdbRecord{}
	bulkCount := 0
	bulkSize := 10000
	bulkId := make([]int64, bulkSize)
	bulkV := make([]float32, config.Db.Faiss.Dimension * bulkSize)
	collectionBulkCount := map[string]int{}
	collectionBulkId := map[string][]int64{}
	collectionBulkV := map[string][]float32{}
	for _, collection := range config.Db.Faiss.Collections {
		collectionBulkCount[collection] = 0
		collectionBulkId[collection] = make([]int64, bulkSize)
		collectionBulkV[collection] = make([]float32, config.Db.Faiss.Dimension * bulkSize)
	}
	it := dataDB.db.NewIterator(dataDB.defaultReadOptions)
	it.Seek([]byte(start))
	defer it.Close()
	for it = it; it.Valid(); it.Next() {
		key := it.Key()
		value := it.Value()
		defer key.Free()
		defer value.Free()
		DecodeFaissdbRecord(tmpFaissdbRecord, value.Data())
		bulkId[bulkCount] = tmpFaissdbRecord.Id
		copy(bulkV[(bulkCount * config.Db.Faiss.Dimension):((bulkCount+1)*config.Db.Faiss.Dimension)], tmpFaissdbRecord.V)
		for _, collection := range tmpFaissdbRecord.Collections {
			if self.indexes[collection] != nil {
				collectionBulkId[collection][collectionBulkCount[collection]] = tmpFaissdbRecord.Id
				copy(collectionBulkV[collection][(collectionBulkCount[collection] * config.Db.Faiss.Dimension):((collectionBulkCount[collection]+1)*config.Db.Faiss.Dimension)], tmpFaissdbRecord.V)
				collectionBulkCount[collection]++
			}
		}
		idDB.PutString(strconv.FormatInt(bulkId[bulkCount], 10), string(key.Data()))
		bulkCount++
		if bulkCount == bulkSize {
			log.Println("bulkAdd start", self.Ntotal())
			idxErr := self.mainIndex.AddWithIDs(bulkV, bulkId)
			if idxErr != nil {
				log.Println(idxErr)
			}
			for _, collection := range config.Db.Faiss.Collections {
				if collectionBulkCount[collection] > 0 {
					idxErr := self.indexes[collection].AddWithIDs(collectionBulkV[collection][0:(collectionBulkCount[collection] * config.Db.Faiss.Dimension)], collectionBulkId[collection][0:collectionBulkCount[collection]])
					if idxErr != nil {
						log.Println(idxErr)
					}
					collectionBulkCount[collection] = 0
					collectionBulkId[collection] = make([]int64, bulkSize)
					collectionBulkV[collection] = make([]float32, config.Db.Faiss.Dimension * bulkSize)
				}
			}
			bulkId = make([]int64, bulkSize)
			bulkV = make([]float32, config.Db.Faiss.Dimension * bulkSize)
			bulkCount = 0
			log.Println("bulkAdd", self.Ntotal())
		}
	}
	if bulkCount > 0 {
		bulkId = bulkId[0:bulkCount]
		bulkV = bulkV[0:(bulkCount*config.Db.Faiss.Dimension)]
		idxErr := self.mainIndex.AddWithIDs(bulkV, bulkId)
		if idxErr != nil {
			log.Println(idxErr)
		}
		for _, collection := range config.Db.Faiss.Collections {
			if collectionBulkCount[collection] > 0 {
				idxErr := self.indexes[collection].AddWithIDs(collectionBulkV[collection][0:(collectionBulkCount[collection]*config.Db.Faiss.Dimension)], collectionBulkId[collection][0:collectionBulkCount[collection]])
				if idxErr != nil {
					log.Println(idxErr)
				}
			}
		}
	}
	self.Write()
}

func (self *LocalIndex) Train(trainData []float32) {
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

func syncThread() {
	for ;; {
		time.Sleep(config.Db.Faiss.Syncinterval * time.Millisecond)
		if FaissdbStatus == STATUS_READY {
			localIndex.Write()
		}
	}
}

func InitLocalIndex() {
	localIndex = newLocalIndex()
	localIndex.Open()
	go syncThread()
}

func GapSyncLocalIndex() {
	lastkey := LastKey()
	metaLastkey := metaDB.GetString("lastkey")
	if lastkey != metaLastkey {
		log.Printf("Detect gap index(%v) != localdb(%v)", metaLastkey, lastkey)
		localIndex.SyncFromLocalDb(metaLastkey)
	}
}
