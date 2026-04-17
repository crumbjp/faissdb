package main

import (
	"gopkg.in/yaml.v2"
	"log"
	"time"
)

var config Config

type Faissconfig struct {
	Description  string
	Metric       string
	Nprobe       int
	Directmap    *bool
	Dimension    int
	Syncinterval time.Duration
}

func (self Faissconfig) UseDirectMap() bool {
	if self.Directmap == nil {
		return true
	}
	return *self.Directmap
}

type Dbconfig struct {
	Capacity uint64
}

type Replicaonfig struct {
	Listen string
}

type Config struct {
	Process struct {
		Performancelog bool
		Loglv          string
		Logfile        string
		Pidfile        string
		Daemon         bool
		Memlimit       int64
	}
	Http struct {
		MaxConnections int
		Port           int
		HttpTimeout    int
	}
	Db struct {
		Dbpath    string
		Faiss     Faissconfig
		Metadb    Dbconfig
		Datadb    Dbconfig
		Iddb      Dbconfig
		Oplogdb   Dbconfig
		Replicadb Dbconfig
	}
	Oplog struct {
		Term int
	}
	Feature struct {
		Listen string
	}
	Replica Replicaonfig
}

func loadConfig(configFile string) {
	data, err := ReadFile(configFile)
	if err != nil {
		log.Fatalf("loadConfig() err1 %v", err)
	}
	config = Config{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("loadConfig() err %v", err)
	}
}
