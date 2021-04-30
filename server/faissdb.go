package main

import (
	"os"
	"errors"
	"sync"
	"log"
	"github.com/sevlyar/go-daemon"
)

const (
	STATUS_NONE = 0
	STATUS_STARTUP = 10
	STATUS_TRAINING = 20
	STATUS_FULLSYNC = 30
	STATUS_READY = 100
	STATUS_TERMINATING = 255

)
var metaDB *LocalDB
var dataDB *LocalDB
var idDB *LocalDB
var rwmutex sync.RWMutex
var FaissdbStatus int
var idGenerator *IdGenerator

func setStatus(status int) error {
	if FaissdbStatus == STATUS_TERMINATING {
		return errors.New("Terminating now")
	}
	FaissdbStatus = status
	return nil
}

func start() {
	log.Println("start()")
	rwmutex = sync.RWMutex{}
	setStatus(STATUS_STARTUP)
	idGenerator = NewIdGenerator()
	metaDB = newLocalDB("/meta")
	metaDB.Open(&config.Db.Metadb)
	dataDB = newLocalDB("/data")
	dataDB.Open(&config.Db.Datadb)
	idDB = newLocalDB("/id")
	idDB.Open(&config.Db.Iddb)
	go InitRpcReplicaServer()
	InitOplog()
	InitRpcReplicaClient()
	if IsPrimary() {
		InitLocalIndex()
		GapSyncLocalIndex()
	} else {
		InitLocalIndex()
		lastKey := LastKey()
		if lastKey == "" {
			ReplicaFullSync()
		} else {
			GapSyncLocalIndex()
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
	setStatus(STATUS_READY)
	go InitRpcFeatureServer()
	InitHttpServer()
}

func main() {
	configFile := "config.yml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	loadConfig(configFile)
	logfile, err := os.OpenFile(config.Process.Logfile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Failure to open logfile %s", config.Process.Logfile)
	}
	log.SetOutput(logfile)
	if config.Process.Daemon {
		context := &daemon.Context{
			PidFileName: config.Process.Pidfile,
			PidFilePerm: 0644,
			WorkDir:     "./",
		}
		child, err := context.Reborn()
		if err != nil {
			log.Fatalln(err)
		}
		if child != nil {
			return
		}
		defer context.Release()
		start()
	} else {
		start()
	}
	log.Println("main() end")
}
