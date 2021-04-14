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
	Ntotal int64
	LastsyncedAt string
	Lastkey string
	Faiss Faissconfig
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
			resp, err := json.Marshal(StatusResult{Istrained: localIndex.IsTrained(), Ntotal: localIndex.Ntotal(), Lastkey: LastKey()})
			if err != nil {
				log.Println(err)
				w.Write([]byte(err.Error()))
			} else {
				w.Write(resp)
			}
		}
		return
	} else if IsTraining() || terminating {
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
		} else if r.URL.Path == "/sync" {
			Sync()
		}
		return
	}
}

func InitHttpServer() {
	http.HandleFunc("/", httpHandler)
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
		oplogDB.Close()
		httpServer.Close()
	}()
	err := httpServer.Serve(limit_listener)
	if err != nil {
		log.Fatalln(err)
	}
}
