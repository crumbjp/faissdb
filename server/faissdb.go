package main

    // dimension: 24000
    // description: IVF128,PQ300

import (
	"os"
	"errors"
	"strconv"
	"math/rand"
	"container/list"
	"fmt"
	"time"
	"sync"
	"log"
	"bytes"
	"encoding/binary"
	"gopkg.in/yaml.v2"
)

type Dbconfig struct {
	Capacity uint64
}

type Replicaonfig struct {
	Listen string
	Master string
}

type Config struct {
	Http struct {
		MaxConnections int
		Port int
		HttpTimeout int
	}
	Db struct {
		Dbpath string
		Faiss struct {
			Dimension int
			Description string
			Metric string
			Syncinterval time.Duration
		}
		Datadb Dbconfig
		Iddb Dbconfig
		Oplogdb Dbconfig
	}
	Oplog struct {
		Term int
	}
	Replica Replicaonfig
}

// ----------- Data -----------
type Data struct {
	id int64
	v []float32
}

func (self *Data) Encode() ([]byte, error) {
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.LittleEndian, self.id)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.LittleEndian, self.v)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (self *Data) Decode(b []byte) error {
	buffer := bytes.NewReader(b)
	err := binary.Read(buffer, binary.LittleEndian, &self.id)
	if err != nil {
		return err
	}
	if len(self.v) != config.Db.Faiss.Dimension {
		self.v = make([]float32, config.Db.Faiss.Dimension)
	}
	err = binary.Read(buffer, binary.LittleEndian, &self.v)
	return nil
}

const (
	STATUS_STARTUP = 0
	STATUS_READY = 1
)
// ----------- Logic -----------
var config Config
var dataDB *LocalDB
var idDB *LocalDB
var rwmutex sync.RWMutex
var terminating bool
var training bool
var status int
var idGenerator *IdGenerator

func setId(key string, data *Data) {
	d := Del(key)
	if d != nil {
		data.id = d.id
	} else {
		data.id = idGenerator.Generate()
	}
}

func SetRaw(key string, data *Data) []byte {
	rwmutex.Lock()
	defer rwmutex.Unlock()
	encoded, err := data.Encode()
	if err != nil {
		panic(err)
	}
	dataDB.Put(key, encoded)
	idDB.PutString(strconv.FormatInt(data.id, 10), key)
	localIndex.AddWithIDs(data.v, []int64{data.id})
	return encoded
}

func Set(key string, v []float32) error {
	data := Data{v: v}
	if(len(data.v) != config.Db.Faiss.Dimension) {
		return errors.New(fmt.Sprintf("Invalid data dimensions expected: %d actual: %d", config.Db.Faiss.Dimension, len(data.v)))
	}
	setId(key, &data)
	encoded := SetRaw(key, &data)
	PutOplog(OP_SET, key, encoded)
	return nil
}

func DelRaw(key string, data *Data) {
	dataDB.Delete(key)
	idDB.Delete(strconv.FormatInt(data.id, 10))
	localIndex.RemoveIDs([]int64{data.id})
}

func Del(key string) *Data {
	rwmutex.Lock()
	defer rwmutex.Unlock()
	value := dataDB.Get(key)
	defer value.Free()
	valueData := value.Data()
	if(valueData != nil) {
		data := Data{}
		data.Decode(valueData)
		DelRaw(key, &data)
		data.v = nil
		encoded, err := data.Encode()
		if err != nil {
			panic(err)
		}
		PutOplog(OP_DEL, key, encoded)
		return &data
	}
	return nil
}

func Sync() {
	log.Println("Sync")
	localIndex.Reset()
	idDB.DestroyDb()
	idDB.Open(config.Db.Iddb)
	bulkSize := 10000
	bulkId := make([]int64, bulkSize)
	bulkV := make([]float32, config.Db.Faiss.Dimension * bulkSize)
	tmpData := Data{}
	bulkCount := 0
	it := dataDB.db.NewIterator(dataDB.defaultReadOptions)
	it.Seek([]byte(""))
	defer it.Close()
	for it = it; it.Valid(); it.Next() {
		key := it.Key()
		value := it.Value()
		defer key.Free()
		defer value.Free()
		tmpData.v = bulkV[(bulkCount * config.Db.Faiss.Dimension):((bulkCount+1)*config.Db.Faiss.Dimension)]
		tmpData.Decode(value.Data())
		bulkId[bulkCount] = tmpData.id // Copy
		idDB.PutString(strconv.FormatInt(bulkId[bulkCount], 10), string(key.Data()))
		bulkCount++
		if bulkCount == bulkSize {
			log.Println("bulkAdd start", localIndex.Ntotal())
			idxErr := localIndex.AddWithIDs(bulkV, bulkId)
			if idxErr != nil {
				log.Println(idxErr)
			}
			bulkId = make([]int64, bulkSize)
			bulkV = make([]float32, config.Db.Faiss.Dimension * bulkSize)
			bulkCount = 0
			log.Println("bulkAdd", localIndex.Ntotal())
		}
	}
	if bulkCount > 0 {
		bulkId = bulkId[0:bulkCount]
		bulkV = bulkV[0:(bulkCount*config.Db.Faiss.Dimension)]
		idxErr := localIndex.AddWithIDs(bulkV, bulkId)
		if idxErr != nil {
			log.Println(idxErr)
		}
	}
	localIndex.Write()
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
	var tmpId int64
	count := 0
	trainData := make([]float32, config.Db.Faiss.Dimension * keys.Len())
	for element := keys.Front(); element != nil; element = element.Next() {
		value := dataDB.Get(element.Value.(string))
		defer value.Free()
		valueData := value.Data()
		v := trainData[(count * config.Db.Faiss.Dimension):((count+1)*config.Db.Faiss.Dimension)]
		buffer := bytes.NewReader(valueData)
		err := binary.Read(buffer, binary.LittleEndian, &tmpId)
		if err != nil {
			panic(err)
		}
		err = binary.Read(buffer, binary.LittleEndian, &v)
		if err != nil {
			panic(err)
		}
		count++
	}
	return trainData
}

func setTrain() {
	training = true
}

func unsetTrain() {
	training = false
}

func IsTraining() bool {
	return training
}

func Train(proportion float32, force bool) error {
	if !force && localIndex.IsTrained() {
		return nil
	}
	setTrain()
	defer unsetTrain()
	log.Println("Build train data")
	trainData := buildTrainData(proportion)
	log.Println(fmt.Sprintf("Train start (%d)", len(trainData) / config.Db.Faiss.Dimension))
	localIndex.Train(trainData)
	log.Println("Write trained index")
	localIndex.WriteTrained()
	Sync()
	log.Println("Train end")
	return nil
}

type SearchResult struct {
	distance float32
	key string
}

func Search(v []float32, n int64) ([]SearchResult) {
	distances, labels := localIndex.Search(v, n)
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

func loadConfig() {
	configFile := "config.yml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	data, err := ReadFile(configFile)
	if err != nil {
    log.Fatalf("loadConfig() %v", err)
	}
	config = Config{}
	confErr := yaml.Unmarshal(data, &config)
	if confErr != nil {
    log.Fatalf("loadConfig() %v", err)
	}
}

func main() {
	rwmutex = sync.RWMutex{}
	terminating = false
	training = false
	status = STATUS_STARTUP

	idGenerator = NewIdGenerator()
	loadConfig()
	dataDB = newLocalDB("/data")
	dataDB.Open(config.Db.Datadb)
	idDB = newLocalDB("/id")
	idDB.Open(config.Db.Iddb)
	go InitRpcReplicaServer()
	InitOplog()
	InitRpcReplicaClient()
	if IsMaster() {
		InitLocalIndex()
	} else {
		lastKey := LastKey()
		if lastKey == "" {
			ReplicaFullSync()
		} else {
			InitLocalIndex()
			masterLastKey, err := RpcReplicaGetLastKey()
			if err != nil {
				log.Fatalf("No master %v", err)
			}
			if masterLastKey != lastKey {
				ReplicaSync()
			}
		}
		go InitReplicaSyncThread()
	}
	status = STATUS_READY
	log.Println("Opened Ntotal:", localIndex.Ntotal())
	InitHttpServer()
}
