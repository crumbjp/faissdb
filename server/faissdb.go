package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"errors"
	"sync"
	"time"
	"github.com/sevlyar/go-daemon"
	"github.com/google/uuid"
	"net/http"
)

const (
	STATUS_NONE = 0
	STATUS_STARTUP = 10
	STATUS_CONFIGURING = 15
	STATUS_TRAINING = 20
	STATUS_FULLSYNC = 30
	STATUS_READY = 100
	STATUS_TERMINATING = 255
)

const (
	CHECK_REPLICASET_INTERVAL = 60000
)

type Faissdb struct {
	firstSync bool
	selfUuid string
	logger *Logger
	metaDB *LocalDB
	dataDB *LocalDB
	idDB *LocalDB
	rwmutex sync.RWMutex
	status int
	prevStatus int
	idGenerator *IdGenerator
	oplogKeyGenerator *IdGenerator
	oplogDB *LocalDB
	httpServer *http.Server
	replicaSet *ReplicaSet
	replicaMembers []*ReplicaMember
	selfMember *ReplicaMember
	primaryMember *ReplicaMember
	secondaryMembers []*ReplicaMember
	lastCheckedAt time.Time
	rsJson string
	rsTs int64
	replicaSyncMutex sync.Mutex
}
var faissdb Faissdb

func setStatus(status int) error {
	if faissdb.status == status {
		return nil
	}
	if faissdb.status == STATUS_TERMINATING {
		return errors.New("setStatus() Terminating now")
	}
	if status == STATUS_CONFIGURING {
		if faissdb.status != STATUS_READY && faissdb.status != STATUS_STARTUP {
			return errors.New(fmt.Sprintf("setStatus() Not ready %v", faissdb.status))
		}
	}
	faissdb.prevStatus = faissdb.status
	faissdb.status = status
	return nil
}

func rollbackStatus() {
	faissdb.status = faissdb.prevStatus
}

func start() {
	faissdb.selfUuid = uuid.New().String()
	faissdb.logger.Debug("start() %s", faissdb.selfUuid)
	faissdb.rwmutex = sync.RWMutex{}
	faissdb.replicaSyncMutex = sync.Mutex{}
	setStatus(STATUS_STARTUP)
	faissdb.idGenerator = NewIdGenerator()
	faissdb.metaDB = newLocalDB("/meta")
	faissdb.metaDB.Open(&config.Db.Metadb)
	faissdb.dataDB = newLocalDB("/data")
	faissdb.dataDB.Open(&config.Db.Datadb)
	faissdb.idDB = newLocalDB("/id")
	faissdb.idDB.Open(&config.Db.Iddb)
	go InitRpcReplicaServer()
	InitOplog()
	InitLocalIndex()
	GapSyncLocalIndex()
	go InitReplicaSyncThread()
	InitReplicaSet()
	go InitRpcFeatureServer()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		faissdb.logger.Info("SIGNAL: %v", sig)
		setStatus(STATUS_TERMINATING)
		// TODO: Obtain internal writelock
		faissdb.logger.Info("Termination start")
		localIndex.Write()
		faissdb.idDB.Close()
		faissdb.dataDB.Close()
		faissdb.oplogDB.Close()
		faissdb.metaDB.Close()
		faissdb.httpServer.Close()
		faissdb.logger.Info("Termination end")
	}()
	InitHttpServer()
}

func main() {
	configFile := "config.yml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	loadConfig(configFile)
	InitLogger(config.Process.Logfile)
	if config.Process.Daemon {
		context := &daemon.Context{
			PidFileName: config.Process.Pidfile,
			PidFilePerm: 0644,
			WorkDir:     "./",
		}
		child, err := context.Reborn()
		if err != nil {
			faissdb.logger.Fatal("%v", err)
		}
		if child != nil {
			return
		}
		defer context.Release()
		start()
	} else {
		start()
	}
	faissdb.logger.Info("main() end")
}
