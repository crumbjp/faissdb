package main

import (
	"sync"
	"local.packages/go-faiss" // "github.com/DataIntelligenceCrew/go-faiss"
	"log"
	"time"
)

const (
	FAISS_INDEX = "/faiss"
	FAISS_TRAINED = "/faiss_trained"
)

var localIndex *LocalIndex

type LocalIndex struct {
	rwmutex sync.RWMutex
	index faiss.Index
	parameterSpace *faiss.ParameterSpace
}

func newLocalIndex() *LocalIndex {
	return &LocalIndex{}
}

func IndexFilePath() string {
	return config.Db.Dbpath + FAISS_INDEX
}

func TrainedFilePath() string {
	return config.Db.Dbpath + FAISS_TRAINED
}

func (self *LocalIndex) Open() {
	self.rwmutex = sync.RWMutex{}
	metric := faiss.MetricInnerProduct
	if config.Db.Faiss.Metric == "InnerProduct" {
		metric = faiss.MetricInnerProduct
	} else if config.Db.Faiss.Metric == "L2" {
		metric = faiss.MetricL2
	} else if config.Db.Faiss.Metric == "L1" {
		metric = faiss.MetricL1
	} else if config.Db.Faiss.Metric == "Linf" {
		metric = faiss.MetricLinf
	} else if config.Db.Faiss.Metric == "Lp" {
		metric = faiss.MetricLp
	} else if config.Db.Faiss.Metric == "Canberra" {
		metric = faiss.MetricCanberra
	} else if config.Db.Faiss.Metric == "BrayCurtis" {
		metric = faiss.MetricBrayCurtis
	} else if config.Db.Faiss.Metric == "JensenShannon" {
		metric = faiss.MetricJensenShannon
	}
	index, err := faiss.ReadIndex(IndexFilePath(), faiss.IoFlagMmap)
	if err != nil {
		log.Println(err)
	}
	if index == nil {
		index, err = faiss.IndexFactory(config.Db.Faiss.Dimension, config.Db.Faiss.Description, metric)
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
	err = self.parameterSpace.SetIndexParameter(self.index, "nprobe", 10)
	if err != nil {
		log.Println(err)
	}
}

func (self *LocalIndex) flush(path string) {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	err := faiss.WriteIndex(self.index, path)
	if err != nil {
		panic(err)
	}
}

func (self *LocalIndex) WriteTrained() {
	self.flush(TrainedFilePath());
}

func (self *LocalIndex) Write() {
	self.flush(IndexFilePath());
}

func (self *LocalIndex) IsTrained() (bool) {
	return self.index.IsTrained()
}

func (self *LocalIndex) Reset() {
	self.index.Reset()
}

func (self *LocalIndex) Train(vector []float32) {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	self.index.Reset()
	err := self.index.Train(vector)
	if err != nil {
		panic(err)
	}
}

func (self *LocalIndex) AddWithIDs(vectors []float32, xids []int64) error {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	err := self.index.AddWithIDs(vectors, xids)
	if err != nil {
		log.Println()
	}
	return err
}

func (self *LocalIndex) RemoveIDs(ids []int64) int {
	selector, err := faiss.NewIDSelectorBatch(ids)
	if err != nil {
		panic(err)
	}
	var n int
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	n, err = self.index.RemoveIDs(selector)
	if err != nil {
		panic(err)
	}
	return n
}

func (self *LocalIndex) Search(vector []float32, n int64) ([]float32, []int64) {
	distances, labels, _ := self.index.Search(vector, n)
	return distances, labels
}

func (self *LocalIndex) Ntotal() int64 {
	return self.index.Ntotal()
}

func syncThread() {
	for ;; {
		time.Sleep(config.Db.Faiss.Syncinterval * time.Millisecond)
		if !IsTraining() && !terminating {
			log.Println("localIndex.Write() start")
			localIndex.Write()
			log.Println("localIndex.Write() end")
		}
	}
}

func InitLocalIndex() {
	localIndex = newLocalIndex()
	localIndex.Open()
	go syncThread()
}
