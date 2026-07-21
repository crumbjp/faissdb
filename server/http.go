package main

import (
	"net"
	"os"
	"runtime"
	"strconv"
	"time"
	"log"
	"fmt"
	"io/ioutil"
	"golang.org/x/net/netutil"
	"net/http"
	"encoding/json"
	//	_ "net/http/pprof"
)

type StatusResult struct {
	Status int
	Istrained bool
	Lastsynced string
	Lastkey string
	DataCount int64
	Faiss Faissconfig
	Ntotal map[string]int64
	ReplicaSet *ReplicaSet
	Primary bool
	Secondary bool
}

type IndexMemoryResult struct {
	Ntotal int64
	InvlistsNlist uint64
	InvlistsUsedBytes uint64
	InvlistsReservedBytes uint64
	InvlistsViewBytes uint64
}

type MemoryResult struct {
	RssBytes uint64
	GoAllocBytes uint64
	GoSysBytes uint64
	Rocksdb map[string]map[string]uint64
	Indexes map[string]*IndexMemoryResult
}

func currentRssBytes() uint64 {
	data, err := os.ReadFile("/proc/self/statm")
	if err != nil {
		return 0
	}
	var pages uint64
	fmt.Sscanf(string(data), "%d %d", &pages, &pages)
	return pages * 4096
}

func writeJson(w http.ResponseWriter, result interface{}) {
	resp, err := json.Marshal(result)
	if err != nil {
		faissdb.logger.Info("writeJson() json.Marshal() %v", err)
		log.Println(err)
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
	}
}

// -----------
/*
   *Get status
     get /
   *Execute train
     post /train
       number-of-train-data
 */
func httpHandler(w http.ResponseWriter, r *http.Request) {
	faissdb.logger.Info("httpHandler() %s %s", r.Method, r.URL.Path)
	if err := beginRequest(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	defer endRequest()
	if r.Method == http.MethodGet {
		if r.URL.Path == "/" {
			searchResult := StatusResult{
				Istrained: localIndex.IsTrained(),
				Faiss: config.Db.Faiss,
				Lastsynced: 	faissdb.metaDB.GetString("lastkey"),
				Lastkey: LastKey(),
				DataCount: faissdb.dataDB.Count(),
				Status: faissdb.status,
				Ntotal: map[string]int64{},
				ReplicaSet: faissdb.replicaSet,
				Primary: IsPrimary(),
				Secondary: IsSecondary(),
			}
			for collection := range localIndex.Indexes() {
				searchResult.Ntotal[collection] = localIndex.Ntotal(collection)
			}
			writeJson(w, searchResult)
		} else if r.URL.Path == "/memory" {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			memoryResult := MemoryResult{
				RssBytes: currentRssBytes(),
				GoAllocBytes: memStats.Alloc,
				GoSysBytes: memStats.Sys,
				Rocksdb: map[string]map[string]uint64{
					"meta": faissdb.metaDB.MemoryUsage(),
					"data": faissdb.dataDB.MemoryUsage(),
					"id": faissdb.idDB.MemoryUsage(),
					"log": faissdb.oplogDB.MemoryUsage(),
					"replica": faissdb.replicaDB.MemoryUsage(),
				},
				Indexes: map[string]*IndexMemoryResult{},
			}
			for collection, index := range localIndex.Indexes() {
				indexMemoryResult := &IndexMemoryResult{Ntotal: index.Ntotal()}
				invlistsMemory := index.InvlistsMemory()
				if invlistsMemory != nil {
					indexMemoryResult.InvlistsNlist = invlistsMemory.Nlist
					indexMemoryResult.InvlistsUsedBytes = invlistsMemory.UsedBytes
					indexMemoryResult.InvlistsReservedBytes = invlistsMemory.ReservedBytes
					indexMemoryResult.InvlistsViewBytes = invlistsMemory.ViewBytes
				}
				memoryResult.Indexes[collection] = indexMemoryResult
			}
			writeJson(w, memoryResult)
		} else if r.URL.Path == "/malloc_info" {
			mallocInfo, err := MallocInfo()
			if err != nil {
				faissdb.logger.Info("httpHandler() MallocInfo() %v", err)
				w.WriteHeader(500)
				return
			}
			w.Header().Set("Content-Type", "application/xml")
			w.Write([]byte(mallocInfo))
		}
		return
	} else if r.Method == http.MethodPut {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			faissdb.logger.Info("httpHandler() ioutil.ReadAll() %v", err)
			w.WriteHeader(500)
			return
		}
		if r.URL.Path == "/replicaset" {
			if !IsPrimary() && faissdb.selfMember != nil {
				faissdb.logger.Info("httpHandler() Not permitted")
				w.WriteHeader(500)
				return
			}
			if err := ResetReplicaSet(true, time.Now().UnixNano(), body) ; err != nil {
				faissdb.logger.Info("httpHandler() ResetReplicaSet() %v", err)
				w.WriteHeader(500)
				return
			}
			w.WriteHeader(200)
		}
	} else if r.Method == http.MethodDelete {
		if r.URL.Path == "/replicaset" {
			if err := ShutdownReplicaSet(); err != nil {
				faissdb.logger.Info("httpHandler() ShutdownReplicaSet() %v", err)
				w.WriteHeader(500)
				return
			}
			w.WriteHeader(200)
			return
		}
	} else if faissdb.status != STATUS_READY {
		w.WriteHeader(400)
		return
	} else if r.Method == http.MethodPost {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			faissdb.logger.Info("httpHandler() ioutil.ReadAll() %v", err)
			w.WriteHeader(500)
			return
		}
		strBody := string(body)
		if r.URL.Path == "/train" {
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
		} else if r.URL.Path == "/fullsync" {
			if err := FullLocalSync(); err != nil {
				w.WriteHeader(500)
				return
			}
			w.WriteHeader(200)
		}
		return
	}
}

func InitHttpServer() {
	http.HandleFunc("/", httpHandler)
	http.HandleFunc("/memory", httpHandler)
	http.HandleFunc("/malloc_info", httpHandler)
	http.HandleFunc("/train", httpHandler)
	http.HandleFunc("/ftrain", httpHandler)
	http.HandleFunc("/fullsync", httpHandler)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Http.Port))
	if err != nil {
		faissdb.logger.Fatal("InitHttpServer() Listen() %v", err)
	}
	limit_listener := netutil.LimitListener(listener, config.Http.MaxConnections)
	faissdb.httpServer = &http.Server{
		ReadTimeout:  time.Duration(config.Http.HttpTimeout) * time.Second,
		WriteTimeout: time.Duration(config.Http.HttpTimeout) * time.Second,
	}
	if err := faissdb.httpServer.Serve(limit_listener); err != nil {
		if err == http.ErrServerClosed {
			<-faissdb.shutdownDone
		} else {
			faissdb.logger.Fatal("InitHttpServer() Serve() %v", err)
		}
	}
}
