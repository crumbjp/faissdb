package main

import (
	"strconv"
	"fmt"
	"errors"
	"math/rand"
	"container/list"
	"log"
	pb "github.com/crumbjp/faissdb/server/grpc_replica"
)

func setId(key string, faissdbRecord *pb.FaissdbRecord) {
	deletedRecord := Del(key)
	if deletedRecord != nil {
		faissdbRecord.Id = deletedRecord.Id
	} else {
		faissdbRecord.Id = idGenerator.Generate()
	}
}

func SetRaw(key string, faissdbRecord *pb.FaissdbRecord) []byte {
	rwmutex.Lock()
	defer rwmutex.Unlock()
	encoded, err := EncodeFaissdbRecord(faissdbRecord)
	if err != nil {
		panic(err)
	}
	dataDB.Put(key, encoded)
	idDB.PutString(strconv.FormatInt(faissdbRecord.Id, 10), key)
	localIndex.Add(faissdbRecord)
	return encoded
}

func Set(key string, v []float32, collections []string) error {
	faissdbRecord := pb.FaissdbRecord{V: v, Collections: collections}
	if(len(faissdbRecord.V) != config.Db.Faiss.Dimension) {
		return errors.New(fmt.Sprintf("Invalid dimensions expected: %d actual: %d", config.Db.Faiss.Dimension, len(faissdbRecord.V)))
	}
	setId(key, &faissdbRecord)
	encoded := SetRaw(key, &faissdbRecord)
	PutOplog(OP_SET, key, encoded)
	return nil
}

func DelRaw(key string, faissdbRecord *pb.FaissdbRecord) {
	dataDB.Delete(key)
	idDB.Delete(strconv.FormatInt(faissdbRecord.Id, 10))
	localIndex.Remove(faissdbRecord.Id)
}

func Del(key string) *pb.FaissdbRecord {
	rwmutex.Lock()
	defer rwmutex.Unlock()
	value := dataDB.Get(key)
	defer value.Free()
	valueData := value.Data()
	if(valueData != nil) {
		faissdbRecord := &pb.FaissdbRecord{}
		DecodeFaissdbRecord(faissdbRecord, valueData)
		DelRaw(key, faissdbRecord)
		faissdbRecord.V = nil
		encoded, err := EncodeFaissdbRecord(faissdbRecord)
		if err != nil {
			panic(err)
		}
		PutOplog(OP_DEL, key, encoded)
		return faissdbRecord
	}
	return nil
}

type SearchResult struct {
	distance float32
	key string
}

func Search(collection string, v []float32, n int64) ([]SearchResult) {
	distances, labels := localIndex.Search(collection, v, n)
	count := 0
	searchResults := make([]SearchResult, len(distances))
	for i := 0 ; i < len(distances); i++ {
		if labels[i] != -1 {
			searchResults[count].distance = distances[i]
			searchResults[count].key = string(idDB.GetString(strconv.FormatInt(labels[i], 10)))
			count++
		}
	}
	return searchResults[0:count]
}

func buildTrainData(proportion float32) ([]float32) {
	keys := list.New()
	dataDB.rwmutex.RLock()
	defer dataDB.rwmutex.RUnlock()
	it := dataDB.db.NewIterator(dataDB.defaultReadOptions)
	it.Seek([]byte(""))
	defer it.Close()
	for it = it; it.Valid(); it.Next() {
		key := it.Key()
		defer key.Free()
		if rand.Float32() < proportion {
			keys.PushBack(string(key.Data()))
		}
	}
	count := 0
	trainData := make([]float32, config.Db.Faiss.Dimension * keys.Len())
	for element := keys.Front(); element != nil; element = element.Next() {
		value := dataDB.Get(element.Value.(string))
		defer value.Free()
		valueData := value.Data()
		v := trainData[(count * config.Db.Faiss.Dimension):((count+1)*config.Db.Faiss.Dimension)]
		tmpRecord := &pb.FaissdbRecord{}
		DecodeFaissdbRecord(tmpRecord, valueData)
		copy(v, tmpRecord.V)
		count++
	}
	return trainData
}


func Train(proportion float32, force bool) error {
	if !force && localIndex.IsTrained() {
		return nil
	}
	err := setStatus(STATUS_TRAINING)
	if err != nil {
		return err
	}
	log.Println("Build train data")
	trainData := buildTrainData(proportion)
	log.Println(fmt.Sprintf("Train start (%d)", len(trainData) / config.Db.Faiss.Dimension))
	localIndex.Train(trainData)
	FullLocalSync()
	setStatus(STATUS_READY)
	log.Println("Train end")
	return nil
}

func FullLocalSync() {
	log.Println("FullLocalSync()")
	localIndex.ResetToTrained()
	idDB.DestroyDb()
	idDB.Open(config.Db.Iddb)
	localIndex.SyncFromLocalDb("")
}
