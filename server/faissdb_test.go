package main

import (
	"fmt"
	"os"
	"sync"
	"testing"
	pb "github.com/crumbjp/faissdb/server/grpc_replica"
	"github.com/stretchr/testify/assert"
)

const (
	TEST_FAISS_INDEX_NAME = "testFaissIndex"
)

var vectors []float32
var vectorIds []int64
var faissIndex *FaissIndex
func TestMain(m *testing.M) {
	fmt.Println("------ TestMain() -------")
	loadConfig("../config/test/config1.yml")
	InitLogger(config.Process.Logfile)
	faissdb.idGenerator = NewIdGenerator()
	faissdb.rwmutex = sync.RWMutex{}
	setStatus(STATUS_STARTUP)
	faissdb.metaDB = newLocalDB("/meta")
	faissdb.metaDB.Open(&config.Db.Metadb)
	faissdb.dataDB = newLocalDB("/data")
	faissdb.dataDB.Open(&config.Db.Datadb)
	faissdb.idDB = newLocalDB("/id")
	faissdb.idDB.Open(&config.Db.Iddb)
	InitOplog()

	faissIndex = newFaissIndex(TEST_FAISS_INDEX_NAME)
	n := 256
	vectors = make([]float32, n * config.Db.Faiss.Dimension)
	vectorIds = make([]int64, n)
	for i := 0; i < n; i++ {
		vectors[i*config.Db.Faiss.Dimension] = float32(i)/float32(n)
		vectors[i*config.Db.Faiss.Dimension+1] = float32(i)/float32(n)
		vectorIds[i] = int64(i)
	}
	code := m.Run()
	os.Exit(code)
}

var localDB *LocalDB
func TestLocaldb_Open(t *testing.T) {
	dbconfig := &Dbconfig{Capacity: 1073741824}
	localDB = newLocalDB("/test")
	localDB.Open(dbconfig)
	localDB.Put("bin", []byte("bin"))
	localDB.Put("bin2", []byte("bin2"))
	localDB.PutString("string", "string")
	localDB.PutInt64("int64", int64(999))
	localDB.Close()
	localDB.Open(dbconfig)
	assert.Equal(t, []byte{0x62, 0x69, 0x6e}, localDB.Get("bin").Data())
	assert.Equal(t, []byte{0x62, 0x69, 0x6e, 0x32}, localDB.Get("bin2").Data())
	assert.Equal(t, "string", localDB.GetString("string"))
	assert.Equal(t, int64(999), *localDB.GetInt64("int64"))
	localDB.Delete("bin2")
	keys, values, next := localDB.GetRawData("", 2)
	assert.Equal(t, []string{"bin", "int64"}, keys)
	assert.Equal(t, 2, len(values))
	assert.Equal(t, []byte{0x62, 0x69, 0x6e}, values[0])
	assert.Equal(t, []byte{0xe7, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, values[1])
	assert.Equal(t, "string", next)
}

func TestFaissIndex_IndexFilePath(t *testing.T) {
	assert.Equal(t, faissIndex.IndexFilePath(), config.Db.Dbpath + "/" + TEST_FAISS_INDEX_NAME)
}

func TestFaissIndex_OpenNew(t *testing.T) {
	assert.Error(t, faissIndex.Open(true))
	assert.Error(t, faissIndex.Open(false))
	faissIndex.OpenNew()
}

func TestFaissIndex_OpenNewAlready(t *testing.T) {
	defer func() {
		recover()
	}()
	faissIndex.OpenNew()
	assert.Equal(t, "Should panic", nil)
}


func TestFaissIndex_OpenAlready(t *testing.T) {
	defer func() {
		recover()
	}()
	assert.Equal(t, "Should panic", faissIndex.Open(false))
}

func TestFaissIndex_CloseOpen(t *testing.T) {
	faissIndex.CloseWithoutLock()
	faissIndex.OpenNew()
}

func TestFaissIndex_AddWithIDs_BeforeTrain(t *testing.T) {
	assert.Error(t, faissIndex.AddWithIDs(vectors, vectorIds))
}

func TestFaissIndex_Train(t *testing.T) {
	faissIndex.Train(vectors)
}

func TestFaissIndex_Write(t *testing.T) {
	faissIndex.WriteTrained()
	faissIndex.CloseWithoutLock()
	assert.NoError(t, faissIndex.Open(true))
}

func TestFaissIndex_AddWithIDs(t *testing.T) {
	assert.NoError(t, faissIndex.AddWithIDs(vectors, vectorIds))
}

func TestFaissIndex_Search(t *testing.T) {
	_, ids := faissIndex.Search([]float32{0.1,0.1}, 2)
	assert.Equal(t, []int64{255, 254}, ids)
}

func TestFaissIndex_RemoveIDs(t *testing.T) {
	assert.Equal(t, faissIndex.RemoveIDs([]int64{256, 255, 254}), 2)
	_, ids := faissIndex.Search([]float32{0.1,0.1}, 2)
	assert.Equal(t, []int64{253, 252}, ids)
}

func TestFaissIndex_Ntotal(t *testing.T) {
	assert.Equal(t, int64(254), faissIndex.Ntotal())
}

func TestLocalIndex_initLocalIndex(t *testing.T) {
	mainIndex := newFaissIndex("main")
	assert.NoError(t, mainIndex.Open(true))
	mainIndex.Write()
	mainIndex.CloseWithoutLock()
	initLocalIndex()
	assert.NoError(t, localIndex.OpenAllIndex())
}

func TestLogic_Set(t *testing.T) {
	assert.Error(t, Set("ignore", []float32{0.1,0.1,0.1}, []string{"main", "foo", "bar"}))
	assert.NoError(t, Set("key1", []float32{0.1,0.1}, []string{"main", "foo", "bar", "baz"}))
	assert.NoError(t, Set("key2", []float32{0.1,0.2}, []string{"main", "bar", "baz"}))
	assert.NoError(t, Set("key3", []float32{0.2,0.2}, []string{"main", "foo", "baz"}))
	assert.NoError(t, Set("key4", []float32{0.2,0.3}, []string{"main", "foo", "bar"}))
	localIndex.Write()
	assert.Equal(t, int64(4), localIndex.Ntotal("main"))
	assert.Equal(t, int64(3), localIndex.Ntotal("foo"))
	assert.Equal(t, int64(3), localIndex.Ntotal("bar"))
	assert.Equal(t, int64(3), localIndex.Ntotal("baz"))
}

func TestLogic_Search(t *testing.T) {
	var searchResults []SearchResult
	searchResults = Search("main", []float32{1,2}, 2) // from main
	assert.Equal(t, []string{"key4", "key3"}, []string{searchResults[0].key, searchResults[1].key})
	searchResults = Search("foo", []float32{1,2}, 2)
	assert.Equal(t, []string{"key4", "key3"}, []string{searchResults[0].key, searchResults[1].key})
	searchResults = Search("bar", []float32{1,2}, 2)
	assert.Equal(t, []string{"key4", "key2"}, []string{searchResults[0].key, searchResults[1].key})
	searchResults = Search("baz", []float32{1,2}, 2)
	assert.Equal(t, []string{"key3", "key2"}, []string{searchResults[0].key, searchResults[1].key})
}

func TestLogic_Del(t *testing.T) {
	Del("key1")
	assert.Equal(t, int64(3), localIndex.Ntotal("main"))
	assert.Equal(t, int64(2), localIndex.Ntotal("foo"))
	assert.Equal(t, int64(2), localIndex.Ntotal("bar"))
	assert.Equal(t, int64(2), localIndex.Ntotal("baz"))
}

func TestLocalIndex_GapSyncLocalIndex(t *testing.T) {
	localIndex.CloseAll()
	assert.NoError(t, localIndex.OpenAllIndex())
	assert.Equal(t, int64(4), localIndex.Ntotal("main"))
	GapSyncLocalIndex()
	assert.Equal(t, int64(3), localIndex.Ntotal("main"))
}

func TestUtil_SubtractStrings(t *testing.T) {
	assert.Equal(t, []string{"a"}, SubtractStrings([]string{"a", "b"}, []string{"b", "c"}))
	assert.Equal(t, []string{}, SubtractStrings([]string{"a", "b"}, []string{"a", "b"}))
	assert.Equal(t, []string{"a", "b"}, SubtractStrings([]string{"a", "b"}, []string{}))
	assert.Equal(t, []string{}, SubtractStrings([]string{}, []string{"a"}))
}

func TestUtil_EqualStringSets(t *testing.T) {
	assert.True(t, EqualStringSets([]string{"a", "b"}, []string{"b", "a"}))
	assert.True(t, EqualStringSets([]string{}, []string{}))
	assert.False(t, EqualStringSets([]string{"a", "b"}, []string{"a"}))
	assert.False(t, EqualStringSets([]string{"a"}, []string{"a", "b"}))
}

func TestLogic_SetCollectionsDiff(t *testing.T) {
	assert.NoError(t, Set("key5", []float32{0.3,0.3}, []string{"main", "foo"}))
	assert.Equal(t, int64(4), localIndex.Ntotal("main"))
	assert.Equal(t, int64(3), localIndex.Ntotal("foo"))
	assert.NoError(t, Set("key5", []float32{0.3,0.3}, []string{"main", "baz"}))
	assert.Equal(t, int64(4), localIndex.Ntotal("main"))
	assert.Equal(t, int64(2), localIndex.Ntotal("foo"))
	assert.Equal(t, int64(3), localIndex.Ntotal("baz"))
	searchResults := Search("baz", []float32{0.3,0.3}, 1)
	assert.Equal(t, "key5", searchResults[0].key)
}

func TestLogic_SetVectorChanged(t *testing.T) {
	assert.NoError(t, Set("key5", []float32{0.4,0.4}, []string{"main", "baz"}))
	assert.Equal(t, int64(4), localIndex.Ntotal("main"))
	assert.Equal(t, int64(3), localIndex.Ntotal("baz"))
	searchResults := Search("main", []float32{0.4,0.4}, 1)
	assert.Equal(t, "key5", searchResults[0].key)
}

func TestLogic_SetRawForce(t *testing.T) {
	value := faissdb.dataDB.Get("key5")
	faissdbRecord := &pb.FaissdbRecord{}
	assert.NoError(t, DecodeFaissdbRecord(faissdbRecord, value.Data()))
	value.Free()
	localIndex.RemoveRaw("main", []int64{faissdbRecord.Id})
	assert.Equal(t, int64(3), localIndex.Ntotal("main"))
	assert.NoError(t, Set("key5", []float32{0.4,0.4}, []string{"main", "baz"}))
	assert.Equal(t, int64(3), localIndex.Ntotal("main"))
	SetRaw("key5", faissdbRecord)
	assert.Equal(t, int64(4), localIndex.Ntotal("main"))
	assert.Equal(t, int64(3), localIndex.Ntotal("baz"))
}

func TestLogic_SetOplogDelta(t *testing.T) {
	assert.NoError(t, Set("key5", []float32{0.4,0.4}, []string{"main", "foo"}))
	value := faissdb.oplogDB.Get(LastKey())
	oplog := &Oplog{}
	assert.NoError(t, oplog.Decode(value.Data()))
	value.Free()
	assert.Equal(t, OP_SET, oplog.op)
	oplogRecord := &pb.FaissdbRecord{}
	assert.NoError(t, DecodeFaissdbRecord(oplogRecord, oplog.d))
	assert.True(t, oplogRecord.Delta)
	assert.Equal(t, []string{"baz"}, oplogRecord.RemovedCollections)
	assert.Equal(t, []string{"foo"}, oplogRecord.AddedCollections)
	assert.Equal(t, []float32{0.4,0.4}, oplogRecord.V)
	dataValue := faissdb.dataDB.Get("key5")
	dataRecord := &pb.FaissdbRecord{}
	assert.NoError(t, DecodeFaissdbRecord(dataRecord, dataValue.Data()))
	dataValue.Free()
	assert.False(t, dataRecord.Delta)
	assert.Equal(t, 0, len(dataRecord.RemovedCollections))
	assert.Equal(t, 0, len(dataRecord.AddedCollections))
	assert.Equal(t, int64(4), localIndex.Ntotal("main"))
	assert.Equal(t, int64(3), localIndex.Ntotal("foo"))
	assert.Equal(t, int64(2), localIndex.Ntotal("baz"))
}

func TestLogic_ApplyOplogDelta(t *testing.T) {
	value := faissdb.dataDB.Get("key5")
	faissdbRecord := &pb.FaissdbRecord{}
	assert.NoError(t, DecodeFaissdbRecord(faissdbRecord, value.Data()))
	value.Free()
	localIndex.RemoveRaw("main", []int64{faissdbRecord.Id})
	assert.Equal(t, int64(3), localIndex.Ntotal("main"))
	faissdbRecord.Collections = []string{"main", "bar"}
	faissdbRecord.Delta = true
	faissdbRecord.RemovedCollections = []string{"foo"}
	faissdbRecord.AddedCollections = []string{"bar"}
	encoded, err := EncodeFaissdbRecord(faissdbRecord)
	assert.NoError(t, err)
	assert.NoError(t, ApplyOplog(&Oplog{op: OP_SET, key: "key5", d: encoded}))
	assert.Equal(t, int64(3), localIndex.Ntotal("main"))
	assert.Equal(t, int64(2), localIndex.Ntotal("foo"))
	assert.Equal(t, int64(3), localIndex.Ntotal("bar"))
}

func TestLogic_ApplyOplogDeltaDiverged(t *testing.T) {
	value := faissdb.dataDB.Get("key5")
	faissdbRecord := &pb.FaissdbRecord{}
	assert.NoError(t, DecodeFaissdbRecord(faissdbRecord, value.Data()))
	value.Free()
	assert.Equal(t, 0, len(faissdbRecord.RemovedCollections))
	faissdbRecord.Delta = true
	faissdbRecord.RemovedCollections = []string{"baz"}
	encoded, err := EncodeFaissdbRecord(faissdbRecord)
	assert.NoError(t, err)
	assert.NoError(t, ApplyOplog(&Oplog{op: OP_SET, key: "key5", d: encoded}))
	assert.Equal(t, int64(4), localIndex.Ntotal("main"))
	assert.Equal(t, int64(3), localIndex.Ntotal("bar"))
	assert.Equal(t, int64(2), localIndex.Ntotal("baz"))
}

func TestLogic_SetCollections(t *testing.T) {
	assert.Error(t, SetCollections("nokey", []string{"main"}))
	assert.NoError(t, SetCollections("key5", []string{"main", "baz"}))
	assert.Equal(t, int64(4), localIndex.Ntotal("main"))
	assert.Equal(t, int64(2), localIndex.Ntotal("bar"))
	assert.Equal(t, int64(3), localIndex.Ntotal("baz"))
	value := faissdb.oplogDB.Get(LastKey())
	oplog := &Oplog{}
	assert.NoError(t, oplog.Decode(value.Data()))
	value.Free()
	oplogRecord := &pb.FaissdbRecord{}
	assert.NoError(t, DecodeFaissdbRecord(oplogRecord, oplog.d))
	assert.True(t, oplogRecord.Delta)
	assert.Equal(t, []string{"bar"}, oplogRecord.RemovedCollections)
	assert.Equal(t, []string{"baz"}, oplogRecord.AddedCollections)
	assert.Equal(t, []float32{0.4,0.4}, oplogRecord.V)
	dataValue := faissdb.dataDB.Get("key5")
	dataRecord := &pb.FaissdbRecord{}
	assert.NoError(t, DecodeFaissdbRecord(dataRecord, dataValue.Data()))
	dataValue.Free()
	assert.False(t, dataRecord.Delta)
	assert.True(t, EqualStringSets([]string{"main", "baz"}, dataRecord.Collections))
	assert.Equal(t, []float32{0.4,0.4}, dataRecord.V)
	searchResults := Search("baz", []float32{0.4,0.4}, 1)
	assert.Equal(t, "key5", searchResults[0].key)
}
