package main

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
)

type Dbconfig struct {
	Capacity uint64
	Dbpath string
}

type Config struct {
	Http struct {
		MaxConnections int
		Port int
		HttpTimeout int
	}
	Faiss struct {
		Dimension int
		Description string
		Metric string
		Dbpath string
	}
	Datadb Dbconfig
	Iddb Dbconfig
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
	self.v = make([]float32, config.Faiss.Dimension)
	err = binary.Read(buffer, binary.LittleEndian, &self.v)
	return nil
}

// ----------- LocalDB -----------
type LocalDB struct {
	rwmutex sync.RWMutex
	name string
	defaultBlockBasedTableOptions *gorocksdb.BlockBasedTableOptions
	defaultOptions *gorocksdb.Options
	db *gorocksdb.DB
	defaultReadOptions *gorocksdb.ReadOptions
	defaultWriteOptions *gorocksdb.WriteOptions
}

func newLocalDB() *LocalDB {
	localDb := &LocalDB{}
	localDb.rwmutex = sync.RWMutex{}
	return localDb
}

func (self *LocalDB) Open(dbconfig Dbconfig) {
	self.name = dbconfig.Dbpath
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

func (self *LocalDB) Get(key string) ([]byte) {
	self.rwmutex.RLock()
	defer self.rwmutex.RUnlock()
	value, err := self.db.Get(self.defaultReadOptions, []byte(key))
	if err != nil {
		panic(err)
	}
	defer value.Free() // @@@ Unsafe ??
	return value.Data()
}

func (self *LocalDB) GetString(key string) string {
	return string(self.Get(key))
}

func (self *LocalDB) GetInt64(key string) *int64 {
	var result int64
	buffer := bytes.NewReader(self.Get(key))
	err := binary.Read(buffer, binary.LittleEndian, &result)
	if err != nil {
		return nil
	}
	return &result
}

// ----------- LocalIndex -----------
type LocalIndex struct {
	rwmutex sync.RWMutex
	index *faiss.Index
	parameterSpace *faiss.ParameterSpace
}

func newLocalIndex() *LocalIndex {
	return &LocalIndex{}
}

func (self *LocalIndex) Open() {
	self.rwmutex = sync.RWMutex{}
	metric := faiss.MetricInnerProduct
	if config.Faiss.Metric == "InnerProduct" {
		metric = faiss.MetricInnerProduct
	} else if config.Faiss.Metric == "L2" {
		metric = faiss.MetricL2
	} else if config.Faiss.Metric == "L1" {
		metric = faiss.MetricL1
	} else if config.Faiss.Metric == "Linf" {
		metric = faiss.MetricLinf
	} else if config.Faiss.Metric == "Lp" {
		metric = faiss.MetricLp
	} else if config.Faiss.Metric == "Canberra" {
		metric = faiss.MetricCanberra
	} else if config.Faiss.Metric == "BrayCurtis" {
		metric = faiss.MetricBrayCurtis
	} else if config.Faiss.Metric == "JensenShannon" {
		metric = faiss.MetricJensenShannon
	}
	index, err := faiss.IndexFactory(config.Faiss.Dimension, config.Faiss.Description, metric)
	if err != nil {
		panic(err)
	}
	self.index = index
	err = self.index.ReadIndex(config.Faiss.Dbpath)
	if err != nil {
		log.Println(err)
	}
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

func (self *LocalIndex) Write() {
	self.rwmutex.Lock()
	defer self.rwmutex.Unlock()
	err := self.index.WriteIndex(config.Faiss.Dbpath)
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
func syncThread() {
	for ;; {
		time.Sleep(60000 * time.Millisecond)
		if !terminating && !training {
			log.Println("localIndex.Write() start")
			localIndex.Write()
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
	if(len(data.v) != config.Faiss.Dimension) {
		return errors.New(fmt.Sprintf("Invalid data dimensions expected: %d actual: %d", config.Faiss.Dimension, len(data.v)))
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
	localIndex.AddWithIDs(data.v, []int64{data.id})
	return nil
}

func Del(key string) *Data {
	var data *Data
	data = nil
	rwmutex.Lock()
	defer rwmutex.Unlock()
	value := dataDB.Get(key)
	if(value != nil) {
		data = &Data{}
		data.Decode(value)
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
	idDB.Open(config.Iddb)
	idDB.PutInt64("current", currentId)
	bulkSize := 10000
	bulkId := make([]int64, bulkSize)
	bulkV := make([]float32, config.Faiss.Dimension * bulkSize)
	bulkCount := 0
	it := dataDB.db.NewIterator(dataDB.defaultReadOptions)
	it.Seek([]byte(""))
	defer it.Close()
	for it = it; it.Valid(); it.Next() {
		key := it.Key()
		value := it.Value()
		v := bulkV[(bulkCount * config.Faiss.Dimension):((bulkCount+1)*config.Faiss.Dimension)]
		buffer := bytes.NewReader(value.Data())
		err := binary.Read(buffer, binary.LittleEndian, &bulkId[bulkCount])
		if err != nil {
			panic(err)
		}
		err = binary.Read(buffer, binary.LittleEndian, &v)
		if err != nil {
			panic(err)
		}
		idDB.PutString(strconv.FormatInt(bulkId[bulkCount], 10), string(key.Data()))
		key.Free()
		value.Free()
		bulkCount++
		if bulkCount == bulkSize {
			log.Println("bulkAdd start", localIndex.Ntotal())
			idxErr := localIndex.AddWithIDs(bulkV, bulkId)
			if idxErr != nil {
				log.Println(idxErr)
			}
			bulkId = make([]int64, bulkSize)
			bulkV = make([]float32, config.Faiss.Dimension * bulkSize)
			bulkCount = 0
			log.Println("bulkAdd", localIndex.Ntotal())
		}
	}
	if bulkCount > 0 {
		bulkId = bulkId[0:bulkCount]
		bulkV = bulkV[0:((bulkCount+1)*config.Faiss.Dimension)]
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
		if rand.Float32() < proportion {
			keys.PushBack(string(key.Data()))
		}
		key.Free()
	}
	var tmpId int64
	count := 0
	trainData := make([]float32, config.Faiss.Dimension * keys.Len())
	for element := keys.Front(); element != nil; element = element.Next() {
		value := dataDB.Get(element.Value.(string))
		v := trainData[(count * config.Faiss.Dimension):((count+1)*config.Faiss.Dimension)]
		buffer := bytes.NewReader(value)
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

func Train(proportion float32, force bool) error {
	if !force && localIndex.IsTrained() {
		return nil
	}
	training = true
	log.Println("Build train data")
	trainData := buildTrainData(proportion)
	log.Println(fmt.Sprintf("Train start (%d)", len(trainData) / config.Faiss.Dimension))
	localIndex.Train(trainData)
	log.Println("Write trained index")
	localIndex.Write()
	Sync()
	training = false
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
	if len(splited) != config.Faiss.Dimension {
		return nil, errors.New("Invalid data")
	}
	v := make([]float32, config.Faiss.Dimension)
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
	v := make([]float32, config.Faiss.Dimension)
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
		if i >= config.Faiss.Dimension {
			return nil, errors.New(fmt.Sprintf("Invalid data dimensions expected: %d actual: %d", config.Faiss.Dimension, i))
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
			lfIndex := strings.Index(strBody, "\n")
			key := strBody[0:lfIndex]
			value := strBody[lfIndex+1:len(strBody)]
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
			w.WriteHeader(200)
			return
		} else if r.URL.Path == "/del" {
			Del(strBody)
			w.WriteHeader(200)
			return
		} else if r.URL.Path == "/search" || r.URL.Path == "/ssearch" {
			lfIndex := strings.Index(strBody, "\n")
			n, err := strconv.Atoi(strBody[0:lfIndex])
			if err != nil {
				log.Println(err)
				w.WriteHeader(500)
				return
			}
			value := strBody[lfIndex+1:len(strBody)]
			var v []float32
			if r.URL.Path == "/search" {
				v, err = parseDenseVector(value)
			} else if r.URL.Path == "/ssearch" {
				v, err = parseSparseVector(value)
			}
			if err != nil {
				log.Println(err)
				w.WriteHeader(500)
				return
			}
			searchResults := Search(v, int64(n))
			resp := ""
			for _, searchResult := range searchResults {
				resp += strconv.FormatFloat(float64(searchResult.distance), 'f', -1, 32) + " " + searchResult.key + "\n"
			}
			w.Write([]byte(resp))
		} else if r.URL.Path == "/train" {
			proportion, err := strconv.ParseFloat(strBody, 32)
			if err != nil {
				w.WriteHeader(500)
				return
			}
			Train(float32(proportion), false)
			w.WriteHeader(200)
		} else if r.URL.Path == "/ftrain" {
			proportion, err := strconv.ParseFloat(strBody, 32)
			if err != nil {
				w.WriteHeader(500)
				return
			}
			Train(float32(proportion), true)
			w.WriteHeader(200)
		} else if r.URL.Path == "/sync" {
			Sync()
		}
		return
	}
}

func loadConfig() {
	config = Config{}
	file, err := os.Open("config.yml")
	if err != nil {
    log.Fatalln(err)
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
    log.Fatalln(err)
	}
	data := make([]byte, fi.Size())
	_, readErr := file.Read(data)
	if readErr != nil {
    log.Fatalln(readErr)
	}
	confErr := yaml.Unmarshal([]byte(data), &config)
	if confErr != nil {
		log.Fatalln(confErr)
	}
}

func main() {
	rwmutex = sync.RWMutex{}
	terminating = false
	training = false
	loadConfig()
	dataDB = newLocalDB()
	dataDB.Open(config.Datadb)
	idDB = newLocalDB()
	idDB.Open(config.Iddb)
	localIndex = newLocalIndex()
	localIndex.Open()
	p := idDB.GetInt64("current")
	if p != nil {
		currentId = *p
	} else {
		currentId = 1
		idDB.PutInt64("current", currentId)
	}
	log.Println("Opened currentId:", currentId, "Ntotal:", localIndex.Ntotal())
	http.HandleFunc("/", httpHandler)
	http.HandleFunc("/set", httpHandler)
	http.HandleFunc("/sset", httpHandler)
	http.HandleFunc("/del", httpHandler)
	http.HandleFunc("/search", httpHandler)
	http.HandleFunc("/ssearch", httpHandler)
	http.HandleFunc("/train", httpHandler)
	http.HandleFunc("/ftrain", httpHandler)
	http.HandleFunc("/sync", httpHandler)
	listener, listenErr := net.Listen("tcp", fmt.Sprintf(":%d", config.Http.Port))
	if listenErr != nil {
		log.Fatalln(listenErr)
	}
	limit_listener := netutil.LimitListener(listener, config.Http.MaxConnections)
	httpServer = &http.Server{
		ReadTimeout:  time.Duration(config.Http.HttpTimeout) * time.Second,
		WriteTimeout: time.Duration(config.Http.HttpTimeout) * time.Second,
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Println("Signal: ", sig)
		terminating = true
		localIndex.Write()
		idDB.Close()
		dataDB.Close()
		httpServer.Close()
	}()
	go syncThread()
	err := httpServer.Serve(limit_listener)
	if err != nil {
		log.Fatalln(err)
	}
}
