package main

import (
	"net"
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
	Faiss Faissconfig
	Ntotal map[string]int64
	ReplicaSet *ReplicaSet
	Primary bool
	Secondary bool
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
	if r.Method == http.MethodGet {
		if r.URL.Path == "/" {
			searchResult := StatusResult{
				Istrained: localIndex.IsTrained(),
				Faiss: config.Db.Faiss,
				Lastsynced: 	faissdb.metaDB.GetString("lastkey"),
				Lastkey: LastKey(),
				Status: faissdb.status,
				Ntotal: map[string]int64{},
				ReplicaSet: faissdb.replicaSet,
				Primary: IsPrimary(),
				Secondary: IsSecondary(),
			}
			searchResult.Ntotal["main"] = localIndex.Ntotal("")
			for collection, _ := range localIndex.indexes {
				searchResult.Ntotal[collection] = localIndex.Ntotal(collection)
			}
			resp, err := json.Marshal(searchResult)
			if err != nil {
				faissdb.logger.Info("httpHandler() json.Marshal() %v", err)
				log.Println(err)
				w.Write([]byte(err.Error()))
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(resp)
			}
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
		faissdb.logger.Fatal("InitHttpServer() Serve() %v", err)
	}
}
