package main

import (
	"os"
	"os/signal"
	"net"
	"strings"
	"strconv"
	"errors"
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
	Currentid int64
	Ntotal int64
	Lastkey string
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
			resp, err := json.Marshal(StatusResult{Istrained: localIndex.IsTrained(), Ntotal: localIndex.Ntotal(), Lastkey: LastKey()})
			if err != nil {
				log.Println(err)
				w.Write([]byte(err.Error()))
			} else {
				w.Write(resp)
			}
		}
		return
	} else if IsTraining() || training {
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


func InitHttpServer() {
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
		oplogDB.Close()
		httpServer.Close()
	}()
	err := httpServer.Serve(limit_listener)
	if err != nil {
		log.Fatalln(err)
	}
}
