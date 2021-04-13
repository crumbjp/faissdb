package main

    // dimension: 24000
    // description: IVF128,PQ300

import (
	"os"
	"os/signal"
	"syscall"
	"errors"
	"strings"
	"strconv"
	"math/rand"
	"container/list"
	"fmt"
	"net"
	"io/ioutil"
	"golang.org/x/net/netutil"
	"net/http"
	"time"
	"sync"
	"log"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"github.com/tecbot/gorocksdb"
	"local.packages/go-faiss"
	// "github.com/DataIntelligenceCrew/go-faiss"
	"google.golang.org/grpc"
	pb "github.com/crumbjp/faissdb/server/rpcreplica"
	"context"
)

type Dbconfig struct {
	Capacity uint64
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
		}
		Datadb Dbconfig
		Iddb Dbconfig
		Oplogdb Dbconfig
	}
	Oplog struct {
		Term int
	}
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

// ----------- LocalDB -----------
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

// ----------- LocalIndex -----------
type LocalIndex struct {
	rwmutex sync.RWMutex
	index faiss.Index
	parameterSpace *faiss.ParameterSpace
}

func newLocalIndex() *LocalIndex {
	return &LocalIndex{}
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
	index, err := faiss.ReadIndex(config.Db.Dbpath + "/faiss", faiss.IoFlagMmap)
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

func (self *LocalIndex) Write(path string) {
	if path == "" {
		path = "/faiss"
	}
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	err := faiss.WriteIndex(self.index, config.Db.Dbpath + path)
	if err != nil {
		panic(err)
	}
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

// ----------- Logic -----------
var config Config
var dataDB *LocalDB
var idDB *LocalDB
var localIndex *LocalIndex
var rwmutex sync.RWMutex
var currentId int64
var httpServer *http.Server
var terminating bool
var training bool
var nowMutex sync.Mutex
var currentNow int64
var currentNowIndex int

func syncThread() {
	for ;; {
		time.Sleep(60000 * time.Millisecond)
		if !terminating && !training {
			log.Println("localIndex.Write() start")
			localIndex.Write("")
			log.Println("localIndex.Write() end")
		}
	}
}

func setId(key string, data *Data) {
	d := Del(key)
	if d != nil {
		data.id = d.id
	} else {
		rwmutex.Lock()
		defer rwmutex.Unlock()
		data.id = currentId
		currentId++
		idDB.PutInt64("current", currentId)
	}
}


func Set(key string, v []float32) error {
	data := Data{v: v}
	if(len(data.v) != config.Db.Faiss.Dimension) {
		return errors.New(fmt.Sprintf("Invalid data dimensions expected: %d actual: %d", config.Db.Faiss.Dimension, len(data.v)))
	}
	setId(key, &data)
	rwmutex.Lock()
	defer rwmutex.Unlock()
	encoded, err := data.Encode()
	if err != nil {
		panic(err)
	}
	dataDB.Put(key, encoded)
	idDB.PutString(strconv.FormatInt(data.id, 10), key)
	logKey := generateLogKey()
	oplog := Oplog{op: OP_SET, d: encoded}
	encodedOplog, _ := oplog.Encode()
	oplogDB.Put(logKey, encodedOplog)
	localIndex.AddWithIDs(data.v, []int64{data.id})
	return nil
}

func Del(key string) *Data {
	var data *Data
	data = nil
	rwmutex.Lock()
	defer rwmutex.Unlock()
	value := dataDB.Get(key)
	defer value.Free()
	valueData := value.Data()
	if(valueData != nil) {
		data = &Data{}
		data.Decode(valueData)
		dataDB.Delete(key)
		idDB.Delete(strconv.FormatInt(data.id, 10))
		localIndex.RemoveIDs([]int64{data.id})
	}
	return data
}

func Sync() {
	log.Println("Sync")
	localIndex.Reset()
	idDB.DestroyDb()
	idDB.Open(config.Db.Iddb)
	idDB.PutInt64("current", currentId)
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
	localIndex.Write("")
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
	localIndex.Write("/faiss_trained")
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
	searchResults := make([]SearchResult, len(distances))
	for i := 0 ; i < len(distances); i++ {
		searchResults[i].distance = distances[i]
		value := idDB.GetString(strconv.FormatInt(labels[i], 10))
		searchResults[i].key = string(value)
	}
	return searchResults
}

type StatusResult struct {
	Istrained bool
	Currentid int64
	Ntotal int64
}

func parseDenseVector(line string) ([]float32, error) {
	splited := strings.Split(line, ",")
	if len(splited) != config.Db.Faiss.Dimension {
		return nil, errors.New("Invalid data")
	}
	v := make([]float32, config.Db.Faiss.Dimension)
	for count, str := range splited {
		f, err := strconv.ParseFloat(str, 32)
		if err != nil {
			return nil, err
		}
		v[count] = float32(f)
	}
	return v, nil
}

func parseSparseVector(line string) ([]float32, error) {
	splited := strings.Split(line, ",")
	v := make([]float32, config.Db.Faiss.Dimension)
	for _, str := range splited {
		colonIndex := strings.Index(str, ":")
		key := str[0:colonIndex]
		value := str[colonIndex+1:len(str)]
		i, err := strconv.Atoi(key)
		if err != nil {
			log.Println("parseSparseVector err", key, err)
			return nil, err
		}
		var f float64
		f, err = strconv.ParseFloat(value, 32)
		if err != nil {
			log.Println("parseSparseVector err", value, err)
			return nil, err
		}
		if i >= config.Db.Faiss.Dimension {
			return nil, errors.New(fmt.Sprintf("Invalid data dimensions expected: %d actual: %d", config.Db.Faiss.Dimension, i))
		}
		v[i] = float32(f)
	}
	return v, nil
}

// -----------
/*
   *Get status
     get /
   *Set data with dense vector
     post /set
       key-string
       v1,v2,....
   *Set data with sparse vector
     post /sset
       key-string
       k1:v1,k2:v2
   *Del data by key
     post /del
       key-string
   *Search with dense vector
     post /search
       number-of-result
       v1,v2,....
   *Search with sparse vector
     post /ssearch
       number-of-result
       k1:v1,k2:v2
   *Execute train
     post /train
       number-of-train-data
 */
func httpHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL.Path)
	if r.Method == http.MethodGet {
		if r.URL.Path == "/" {
			resp, err := json.Marshal(StatusResult{Istrained: localIndex.IsTrained(), Currentid: currentId, Ntotal: localIndex.Ntotal()})
			if err != nil {
				log.Println(err)
				w.Write([]byte(err.Error()))
			} else {
				w.Write(resp)
			}
		}
		return
	} else if terminating || training {
		w.WriteHeader(400)
		return
	} else if r.Method == http.MethodPost {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}
		strBody := string(body)
		if r.URL.Path == "/set" || r.URL.Path == "/sset" {
			var lfIndex int
			var key string
			var value string
			for ;len(strBody) > 0; {
				lfIndex = strings.Index(strBody, "\n")
				key = strBody[0:lfIndex]
				strBody = strBody[lfIndex+1:len(strBody)]
				lfIndex = strings.Index(strBody, "\n")
				if lfIndex < 0 {
					value = strBody
					strBody = ""
				} else {
					value = strBody[0:lfIndex]
					strBody = strBody[lfIndex+1:len(strBody)]
				}
				var v []float32
				var err error
				if r.URL.Path == "/set" {
					v, err = parseDenseVector(value)
				} else if r.URL.Path == "/sset" {
					v, err = parseSparseVector(value)
				}
				if err != nil {
					log.Println(err)
					w.WriteHeader(500)
					return
				}
				err = Set(key, v)
				if err != nil {
					log.Println(err)
					w.WriteHeader(500)
					return
				}
			}
			w.WriteHeader(200)
			return
		} else if r.URL.Path == "/del" {
			var lfIndex int
			var key string
			for ;len(strBody) > 0; {
				lfIndex = strings.Index(strBody, "\n")
				if lfIndex < 0 {
					key = strBody
					strBody = ""
				} else {
					key = strBody[0:lfIndex]
					strBody = strBody[lfIndex+1:len(strBody)]
				}
				Del(key)
			}
			w.WriteHeader(200)
			return
		} else if r.URL.Path == "/search" || r.URL.Path == "/ssearch" {
			lfIndex := strings.Index(strBody, "\n")nv.Atoi(strBody[0:lfIndex])
			if err != nil {
				log.Println(:len(strBody)]
			var v []float32
			if r.URL.Path == "/search" {
				v, err = parseDenseVector(value)
			} else if r.URL.Path == "/ssearch" {
				v, err = parseSparseVector(value)
			}
			if err != nil {
				log.Println(err)
				w.WriteHeader(500)
			 Search(v, int64(n))
			resp := ""
			for _, searchResult := rav.FormatFloat(float64(searchResult.distance), 'f', -1, 32) + " rain(float32(proportion), false)
			w.WriteHeader(200)
		} elseion, err := strconv.ParseFloat(strBody, 32)
			if err != nil {
				w.WriteHeader(500)
				return
			}
			Train(float32(proportion), true)
			w.WriteHeader(200)
		} else if r.URL.Path == "/sfig{}
	file, err := os.Open(configFile)
	if err != nil {
    lofi, err := file.Stat()
	if err != nil {
    log.Fatalln(err)
	}Err := file.Read(data)
	if readErr != nil {
    log.Fatalln(reas *server) SayHello(ctx context.Context, in *pb.HelloRequest) (n.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetNt.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failedync.Mutex{}
	terminating = false
	training = false
	loadConfig(pen(config.Db.Datadb)
	idDB = newLocalDB("/id")
	idDB.Open(confx()
	localIndex.Open()
	intOplog()
	p := idDB.GetInt64("current"/", httpHandler)
	http.HandleFunc("/set", httpHandler)
	http.H.HandleFunc("/del", httpHandler)
	http.HandleFunc("/search", htch", httpHandler)
	http.HandleFunc("/train", httpHandler)
	httpn(listenErr)
	}
	limit_listener := netutil.LimitListener(listen = &http.Server{
		ReadTimeout:  time.Duration(config.Http.Http: time.Duration(config.Http.HttpTimeout) * time.Second,
	}
	sigfy(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig :=pServer.Serve(limit_listener)
	if err != nil {
		log.Fatalln(err)
}
}

