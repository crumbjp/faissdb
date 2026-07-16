package main

import (
	"strconv"
	"fmt"
	"errors"
	"math/rand"
	"slices"
	"container/list"
	pb "github.com/crumbjp/faissdb/server/grpc_replica"
)


func stripDelta(faissdbRecord *pb.FaissdbRecord) {
	faissdbRecord.Delta = false
	faissdbRecord.RemovedCollections = nil
	faissdbRecord.AddedCollections = nil
}

func setUnsafe(key string, faissdbRecord *pb.FaissdbRecord, force bool, currentFaissdbRecord *pb.FaissdbRecord) ([]byte, []string, []string) {
	if currentFaissdbRecord == nil {
		value := faissdb.dataDB.Get(key)
		defer value.Free()
		valueData := value.Data()
		if(valueData != nil) {
			currentFaissdbRecord = &pb.FaissdbRecord{}
			DecodeFaissdbRecord(currentFaissdbRecord, valueData)
		}
	}
	currentV := []float32{}
	currentCollections := []string{}
	if currentFaissdbRecord != nil {
		faissdbRecord.Id = currentFaissdbRecord.Id
		currentV = currentFaissdbRecord.V
		currentCollections = currentFaissdbRecord.Collections
	} else if faissdbRecord.Id == 0 {
		faissdbRecord.Id = faissdb.idGenerator.Generate()
	} else {
	}
	removeCollections := currentCollections
	addCollections := faissdbRecord.Collections
	if !force {
		if faissdbRecord.Delta {
			previousCollections := append(SubtractStrings(faissdbRecord.Collections, faissdbRecord.AddedCollections), faissdbRecord.RemovedCollections...)
			if EqualStringSets(previousCollections, currentCollections) {
				removeCollections = faissdbRecord.RemovedCollections
				addCollections = faissdbRecord.AddedCollections
			}
		} else {
			if slices.Equal(currentV, faissdbRecord.V) {
				removeCollections = SubtractStrings(currentCollections, faissdbRecord.Collections)
				addCollections = SubtractStrings(faissdbRecord.Collections, currentCollections)
			}
		}
	}
	if faissdbRecord.Delta {
		stripDelta(faissdbRecord)
	}
	for _, collection := range removeCollections {
		localIndex.RemoveRaw(collection, []int64{faissdbRecord.Id})
	}
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
	for _, collection := range addCollections {
		localIndex.AddRaw(collection, faissdbRecord)
	}
	faissdb.logger.PerformEnd("SetRaw localIndex", performLocalIndex)
	return encoded, removeCollections, addCollections
}

func SetRaw(key string, faissdbRecord *pb.FaissdbRecord) []byte {
	faissdb.rwmutex.Lock()
	defer faissdb.rwmutex.Unlock()
	encoded, _, _ := setUnsafe(key, faissdbRecord, true, nil)
	return encoded
}

func SetDeltaRaw(key string, faissdbRecord *pb.FaissdbRecord) []byte {
	faissdb.rwmutex.Lock()
	defer faissdb.rwmutex.Unlock()
	encoded, _, _ := setUnsafe(key, faissdbRecord, false, nil)
	return encoded
}

func SyncRaw(key string, faissdbRecord *pb.FaissdbRecord) {
	faissdb.rwmutex.Lock()
	defer faissdb.rwmutex.Unlock()
	faissdb.idDB.PutString(strconv.FormatInt(faissdbRecord.Id, 10), key)
	localIndex.Add(faissdbRecord)
}

func setWithOplogUnsafe(key string, faissdbRecord *pb.FaissdbRecord, currentFaissdbRecord *pb.FaissdbRecord) {
	encoded, removedCollections, addedCollections := setUnsafe(key, faissdbRecord, false, currentFaissdbRecord)
	deltaRecord := &pb.FaissdbRecord{Delta: true, RemovedCollections: removedCollections, AddedCollections: addedCollections}
	encodedDelta, err := EncodeFaissdbRecord(deltaRecord)
	if err != nil {
		panic(err)
	}
	PutOplog(OP_SET, key, append(encoded, encodedDelta...))
}

func Set(key string, v []float32, collections []string) error {
	faissdbRecord := pb.FaissdbRecord{V: v, Collections: Uniq(collections)}
	if(len(faissdbRecord.V) != config.Db.Faiss.Dimension) {
		return errors.New(fmt.Sprintf("Set() Invalid dimensions expected: %d actual: %d", config.Db.Faiss.Dimension, len(faissdbRecord.V)))
	}
	faissdb.rwmutex.Lock()
	defer faissdb.rwmutex.Unlock()
	setWithOplogUnsafe(key, &faissdbRecord, nil)
	return nil
}

func SetCollections(key string, collections []string) error {
	faissdb.rwmutex.Lock()
	defer faissdb.rwmutex.Unlock()
	value := faissdb.dataDB.Get(key)
	defer value.Free()
	valueData := value.Data()
	if valueData == nil {
		return errors.New(fmt.Sprintf("SetCollections() Not found: %s", key))
	}
	currentFaissdbRecord := &pb.FaissdbRecord{}
	DecodeFaissdbRecord(currentFaissdbRecord, valueData)
	faissdbRecord := pb.FaissdbRecord{Id: currentFaissdbRecord.Id, V: currentFaissdbRecord.V, Collections: Uniq(collections)}
	setWithOplogUnsafe(key, &faissdbRecord, currentFaissdbRecord)
	return nil
}

func delUnsafe(key string, faissdbRecord *pb.FaissdbRecord) {
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

func DelRaw(key string, faissdbRecord *pb.FaissdbRecord) {
	faissdb.rwmutex.Lock()
	defer faissdb.rwmutex.Unlock()
	delUnsafe(key, faissdbRecord)
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
		delUnsafe(key, faissdbRecord)
		faissdbRecord.V = nil
		stripDelta(faissdbRecord)
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
			key := faissdb.idDB.GetString(strconv.FormatInt(labels[i], 10))
			if key == "" {
				continue
			}
			searchResults[count].distance = distances[i]
			searchResults[count].key = key
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
	scanned := 0
	for it = it; it.Valid(); it.Next() {
		if scanned % 10000 == 0 {
			faissdb.logger.InfoMem("buildTrainData(%f) scanned %d keys, selected %d", proportion, scanned, keys.Len())
		}
		key := it.Key()
		if rand.Float32() < proportion {
			keys.PushBack(string(key.Data()))
		}
		key.Free()
		scanned++
	}
	faissdb.logger.InfoMem("buildTrainData(%f) scan complete: scanned %d, selected %d", proportion, scanned, keys.Len())
	count := 0
	allocBytes := config.Db.Faiss.Dimension * keys.Len() * 4
	faissdb.logger.Info("buildTrainData(%f) allocating trainData: %d bytes (%dMB)", proportion, allocBytes, allocBytes / 1024 / 1024)
	trainData := make([]float32, config.Db.Faiss.Dimension * keys.Len())
	faissdb.logger.InfoMem("buildTrainData(%f) trainData allocated", proportion)
	tmpRecord := &pb.FaissdbRecord{}
	for element := keys.Front(); element != nil; element = element.Next() {
		if count % 10000 == 0 {
			faissdb.logger.InfoMem("buildTrainData(%f) loaded %d", proportion, count)
		}
		value := faissdb.dataDB.Get(element.Value.(string))
		valueData := value.Data()
		v := trainData[(count * config.Db.Faiss.Dimension):((count+1)*config.Db.Faiss.Dimension)]
		tmpRecord.Reset()
		DecodeFaissdbRecord(tmpRecord, valueData)
		copy(v, tmpRecord.V)
		value.Free()
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
	faissdb.logger.InfoMem("Train(%f) Build data (memlimit=%d)", proportion, config.Process.Memlimit)
	trainData := buildTrainData(proportion)
	faissdb.logger.InfoMem("Train() Train start (%d)", len(trainData) / config.Db.Faiss.Dimension)
	localIndex.Train(trainData)
	faissdb.logger.InfoMem("Train() Train done")
	if err := FullLocalSync(); err != nil {
		faissdb.logger.Error("Train err", err)
		return err
	}
	faissdb.logger.InfoMem("Train end")
	return nil
}

func FullLocalSync() error {
	faissdb.logger.InfoMem("FullLocalSync() start")
	defer faissdb.logger.Info("FullLocalSync() end")
	if err := setStatus(STATUS_FULLSYNC); err != nil {
		return err
	}
	localIndex.ResetToTrained()
	faissdb.logger.InfoMem("FullLocalSync() ResetToTrained done")
	faissdb.idDB.DestroyDb()
	faissdb.idDB.Open(&config.Db.Iddb)
	faissdb.logger.InfoMem("FullLocalSync() idDB reopened")
	localIndex.SyncFromLocalDb()
	faissdb.logger.InfoMem("FullLocalSync() SyncFromLocalDb done")
	PutOplog(OP_FULLSYNC, "", nil)
	if err := setStatus(STATUS_READY); err != nil {
		return err
	}
	return nil
}

func dropallUnsafe() {
	localIndex.ResetToTrained()
	faissdb.idDB.DestroyDb()
	faissdb.idDB.Open(&config.Db.Iddb)
	faissdb.dataDB.DestroyDb()
	faissdb.dataDB.Open(&config.Db.Iddb)
}

func DropallRaw() {
	faissdb.rwmutex.Lock()
	defer faissdb.rwmutex.Unlock()
	dropallUnsafe()
}

func Dropall() error {
	faissdb.logger.InfoMem("Dropall()")
	defer faissdb.logger.Info("Dropall() end")
	faissdb.rwmutex.Lock()
	defer faissdb.rwmutex.Unlock()
	dropallUnsafe()
	PutOplog(OP_DROPALL, "", nil)
	return nil
}

type DbStatsResult struct {
	Status int
	Istrained bool
	Lastsynced string
	Lastkey string
	DataCount int64
	Faiss Faissconfig
	Ntotal map[string]int64
}

func DbStats() DbStatsResult {
	faissdb.logger.InfoMem("DbStats()")
	defer faissdb.logger.Info("DbStats() end")
	dbStatsResult := DbStatsResult{
		Istrained: localIndex.IsTrained(),
		Faiss: config.Db.Faiss,
		Lastsynced: 	faissdb.metaDB.GetString("lastkey"),
		Lastkey: LastKey(),
		DataCount: faissdb.dataDB.Count(),
		Status: faissdb.status,
		Ntotal: map[string]int64{},
	}
	for collection := range localIndex.Indexes() {
		dbStatsResult.Ntotal[collection] = localIndex.Ntotal(collection)
	}
	return dbStatsResult
}
