package main

import (
	"strconv"
	"fmt"
	"errors"
	"math/rand"
	"container/list"
	pb "github.com/crumbjp/faissdb/server/grpc_replica"
)

func setId(key string, faissdbRecord *pb.FaissdbRecord) {
	deletedRecord := Del(key)
	if deletedRecord != nil {
		faissdbRecord.Id = deletedRecord.Id
	} else {
		faissdbRecord.Id = faissdb.idGenerator.Generate()
	}
}

func SetRaw(key string, faissdbRecord *pb.FaissdbRecord) []byte {
	faissdb.rwmutex.Lock()
	defer faissdb.rwmutex.Unlock()
	performEncodeFaissdbRecord := faissdb.logger.PerformStart("SetRaw EncodeFaissdbRecord")
	encoded, err := EncodeFaissdbRecord(faissdbRecord)
	faissdb.logger.PerformEnd("SetRaw EncodeFaissdbRecord", performEncodeFaissdbRecord)
	if err != nil {
		panic(err)
	}
	performDataDB := faissdb.logger.PerformStart("SetRaw dataDB")
	faissdb.dataDB.Put(key, encoded)
	faissdb.logger.PerformEnd("SetRaw dataDB", performDataDB)
	performIdDB := faissdb.logger.PerformStart("SetRaw idDB")
	faissdb.idDB.PutString(strconv.FormatInt(faissdbRecord.Id, 10), key)
	faissdb.logger.PerformEnd("SetRaw idDB", performIdDB)
	performLocalIndex := faissdb.logger.PerformStart("SetRaw localIndex")
	localIndex.Add(faissdbRecord)
	faissdb.logger.PerformEnd("SetRaw localIndex", performLocalIndex)
	return encoded
}

func Set(key string, v []float32, collections []string) error {
	faissdbRecord := pb.FaissdbRecord{V: v, Collections: Uniq(collections)}
	if(len(faissdbRecord.V) != config.Db.Faiss.Dimension) {
		return errors.New(fmt.Sprintf("Set() Invalid dimensions expected: %d actual: %d", config.Db.Faiss.Dimension, len(faissdbRecord.V)))
	}
	setId(key, &faissdbRecord)
	encoded := SetRaw(key, &faissdbRecord)
	PutOplog(OP_SET, key, encoded)
	return nil
}

func DelRaw(key string, faissdbRecord *pb.FaissdbRecord) {
	performDataDB := faissdb.logger.PerformStart("DelRaw dataDB")
	faissdb.dataDB.Delete(key)
	faissdb.logger.PerformEnd("DelRaw dataDB", performDataDB)
	performIdDB := faissdb.logger.PerformStart("DelRaw idDB")
	faissdb.idDB.Delete(strconv.FormatInt(faissdbRecord.Id, 10))
	faissdb.logger.PerformEnd("DelRaw idDB", performIdDB)
	performLocalIndex := faissdb.logger.PerformStart("DelRaw localIndex")
	localIndex.Remove(faissdbRecord)
	faissdb.logger.PerformEnd("DelRaw localIndex", performLocalIndex)
}

func Del(key string) *pb.FaissdbRecord {
	faissdb.rwmutex.Lock()
	defer faissdb.rwmutex.Unlock()
	value := faissdb.dataDB.Get(key)
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
			searchResults[count].key = string(faissdb.idDB.GetString(strconv.FormatInt(labels[i], 10)))
			count++
		}
	}
	return searchResults[0:count]
}

func buildTrainData(proportion float32) ([]float32) {
	keys := list.New()
	faissdb.dataDB.rwmutex.RLock()
	defer faissdb.dataDB.rwmutex.RUnlock()
	it := faissdb.dataDB.db.NewIterator(faissdb.dataDB.defaultReadOptions)
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
		value := faissdb.dataDB.Get(element.Value.(string))
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
	if err := setStatus(STATUS_TRAINING); err != nil {
		return err
	}
	faissdb.logger.Info("Train() Build data")
	trainData := buildTrainData(proportion)
	faissdb.logger.Info("Train() Train start (%d)", len(trainData) / config.Db.Faiss.Dimension)
	localIndex.Train(trainData)
	if err := FullLocalSync(); err != nil {
		return err
	}
	faissdb.logger.Info("Train end")
	return nil
}

func FullLocalSync() error {
	faissdb.logger.Info("FullLocalSync() start")
	defer faissdb.logger.Info("FullLocalSync() end")
	if err := setStatus(STATUS_FULLSYNC); err != nil {
		return err
	}
	localIndex.ResetToTrained()
	faissdb.idDB.DestroyDb()
	faissdb.idDB.Open(&config.Db.Iddb)
	localIndex.SyncFromLocalDb("")
	if err := setStatus(STATUS_READY); err != nil {
		return err
	}
	return nil
}

func DropallRaw() {
	localIndex.ResetToTrained()
	faissdb.idDB.DestroyDb()
	faissdb.idDB.Open(&config.Db.Iddb)
	faissdb.dataDB.DestroyDb()
	faissdb.dataDB.Open(&config.Db.Iddb)
}

func Dropall() error {
	faissdb.logger.Info("Dropall()")
	defer faissdb.logger.Info("Dropall() end")
	DropallRaw()
	PutOplog(OP_DROPALL, "", nil)
	return nil
}

type DbStatsResult struct {
	Status int
	Istrained bool
	Lastsynced string
	Lastkey string
	Faiss Faissconfig
	Ntotal map[string]int64
}

func DbStats() DbStatsResult {
	faissdb.logger.Info("DbStats()")
	defer faissdb.logger.Info("DbStats() end")
	dbStatsResult := DbStatsResult{
		Istrained: localIndex.IsTrained(),
		Faiss: config.Db.Faiss,
		Lastsynced: 	faissdb.metaDB.GetString("lastkey"),
		Lastkey: LastKey(),
		Status: faissdb.status,
		Ntotal: map[string]int64{},
	}
	dbStatsResult.Ntotal["main"] = localIndex.Ntotal("")
	for collection, _ := range localIndex.indexes {
		dbStatsResult.Ntotal[collection] = localIndex.Ntotal(collection)
	}
	return dbStatsResult
}
