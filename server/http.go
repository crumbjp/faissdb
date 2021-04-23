package main

import (
	"os"
	"os/signal"
	"net"
	"strconv"
	"time"
	"log"
	"fmt"
	"syscall"
	"io/ioutil"
	"golang.org/x/net/netutil"
	"net/http"
	"encoding/json"
)

var httpServer *http.Server

type StatusResult struct {
	Istrained bool
	Lastsynced string
	Lastkey string
	Faiss Faissconfig
	Ntotal map[string]int64
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
	log.Println(r.Method, r.URL.Path)
	if r.Method == http.MethodGet {
		if r.URL.Path == "/" {
			searchResult := StatusResult{
				Istrained: localIndex.IsTrained(),
				Faiss: config.Db.Faiss,
				Lastsynced: 	metaDB.GetString("lastkey"),
				Lastkey: LastKey(),
				Ntotal: map[string]int64{},
			}
			searchResult.Ntotal["main"] = localIndex.Ntotal("")
			for _, collection := range config.Db.Faiss.Collections {
				searchResult.Ntotal[collection] = localIndex.Ntotal(collection)
			}
			resp, err := json.Marshal(searchResult)
			if err != nil {
				log.Println(err)
				w.Write([]byte(err.Error()))
			} else {
				w.Write(resp)
			}
		}
		return
	} else if FaissdbStatus != STATUS_READY {
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
			err := setStatus(STATUS_FULLSYNC)
			if err != nil {
				w.WriteHeader(500)
				return
			}
			FullLocalSync()
			setStatus(STATUS_READY)
		}
		return
	}
}

func InitHttpServer() {
	http.HandleFunc("/", httpHandler)
	http.HandleFunc("/train", httpHandler)
	http.HandleFunc("/ftrain", httpHandler)
	http.HandleFunc("/fullsync", httpHandler)
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
		setStatus(STATUS_TERMINATING)
		localIndex.Write()
		idDB.Close()
		dataDB.Close()
		oplogDB.Close()
		httpServer.Close()
	}()
	err := httpServer.Serve(limit_listener)
	if err != nil {
		log.Fatalln(err)
	}
}
