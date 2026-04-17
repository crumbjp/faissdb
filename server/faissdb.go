package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"errors"
	"sync"
	"time"
	"github.com/sevlyar/go-daemon"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	featureServer *grpc.Server
	replicaServer *grpc.Server
	replicaSet *ReplicaSet
	replicaMembers map[int]*ReplicaMember
	selfMember *ReplicaMember
	primaryMember *ReplicaMember
	secondaryMembers []*ReplicaMember
	lastCheckedAt time.Time
	rsJson string
	rsTs int64
	replicaSyncMutex sync.Mutex
	inflightRequests sync.WaitGroup
	terminationOnce sync.Once
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

func beginRequest() error {
	if faissdb.status == STATUS_TERMINATING {
		return errors.New("beginRequest() terminating")
	}
	faissdb.inflightRequests.Add(1)
	if faissdb.status == STATUS_TERMINATING {
		faissdb.inflightRequests.Done()
		return errors.New("beginRequest() terminating")
	}
	return nil
}

func endRequest() {
	faissdb.inflightRequests.Done()
}

func grpcRequestInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := beginRequest(); err != nil {
			return nil, status.Error(codes.Unavailable, "terminating")
		}
		defer endRequest()
		return handler(ctx, req)
	}
}

func beginTermination() error {
	if err := setStatus(STATUS_TERMINATING); err != nil {
		return err
	}
	return nil
}

func shutdownProcess(clearReplicaSet bool) {
	faissdb.terminationOnce.Do(func() {
		faissdb.logger.Info("shutdownProcess() start")
		ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
		defer cancel()
		grpcDone := make(chan struct{})
		go func() {
			var gracefulStopWaitGroup sync.WaitGroup
			if faissdb.featureServer != nil {
				gracefulStopWaitGroup.Add(1)
				go func() {
					defer gracefulStopWaitGroup.Done()
					faissdb.featureServer.GracefulStop()
				}()
			}
			if faissdb.replicaServer != nil {
				gracefulStopWaitGroup.Add(1)
				go func() {
					defer gracefulStopWaitGroup.Done()
					faissdb.replicaServer.GracefulStop()
				}()
			}
			gracefulStopWaitGroup.Wait()
			close(grpcDone)
		}()
		select {
		case <-grpcDone:
		case <-ctx.Done():
			faissdb.logger.Warn("shutdownProcess() grpc graceful stop timeout")
			if faissdb.featureServer != nil {
				faissdb.featureServer.Stop()
			}
			if faissdb.replicaServer != nil {
				faissdb.replicaServer.Stop()
			}
		}
		if faissdb.httpServer != nil {
			if err := faissdb.httpServer.Close(); err != nil {
				faissdb.logger.Error("shutdownProcess() http close %v", err)
			}
		}
		faissdb.inflightRequests.Wait()
		faissdb.replicaSyncMutex.Lock()
		lastKey := LastKey()
		localIndex.Write()
		faissdb.metaDB.PutString("lastkey", lastKey)
		if clearReplicaSet {
			faissdb.metaDB.PutInt64("ReplicaSetTs", 0)
			faissdb.metaDB.PutString("ReplicaSet", "")
		}
		faissdb.idDB.Close()
		faissdb.dataDB.Close()
		faissdb.oplogDB.Close()
		faissdb.metaDB.Close()
		faissdb.replicaSyncMutex.Unlock()
		faissdb.logger.Info("shutdownProcess() end")
		os.Exit(0)
	})
}

func start(fullsyncMode bool) {
	faissdb.selfUuid = uuid.New().String()
	faissdb.logger.InfoMem("start() %s", faissdb.selfUuid)
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
	InitOplog()
	InitLocalIndex()
	GapSyncLocalIndex()
	InitReplicaSet()
	if fullsyncMode {
		faissdb.logger.Info("start() --fullsync: running FullLocalSync")
		if err := FullLocalSync(); err != nil {
			faissdb.logger.Fatal("start() --fullsync: FullLocalSync() %v", err)
		}
		faissdb.logger.Info("start() --fullsync: done, shutting down")
		faissdb.idDB.Close()
		faissdb.dataDB.Close()
		faissdb.oplogDB.Close()
		faissdb.metaDB.Close()
		os.Exit(0)
	}
	go InitRpcReplicaServer()
	go InitReplicaSyncThread()
	go InitRpcFeatureServer()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		faissdb.logger.Info("SIGNAL: %v", sig)
		if err := beginTermination(); err != nil {
			faissdb.logger.Error("beginTermination() %v", err)
		}
		shutdownProcess(false)
	}()
	InitHttpServer()
}

func main() {
	configFile := "config.yml"
	fullsyncMode := false
	for _, arg := range os.Args[1:] {
		if arg == "--fullsync" {
			fullsyncMode = true
			continue
		}
		configFile = arg
	}
	loadConfig(configFile)
	if config.Process.Memlimit > 0 {
		debug.SetMemoryLimit(config.Process.Memlimit)
	}
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
		start(fullsyncMode)
	} else {
		start(fullsyncMode)
	}
	faissdb.logger.Info("main() end")
}
