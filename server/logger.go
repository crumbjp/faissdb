package main

import (
	"io"
	"os"
	"log"
	"time"
)

const (
	LOGLV_DEBUG = 1000000
	LOGLV_TRACE = 100000
	LOGLV_INFO = 10000
	LOGLV_WARN = 1000
	LOGLV_ERROR = 100
	LOGLV_FATAL = 10
)

type PerformContext struct {
	count int64
	elapsed int64
}

type Logger struct {
	loglv int
	performByKey map[string]*PerformContext
	enablePerform bool
}

func (self *Logger) Debug(format string, args ...interface{}) {
	if self.loglv >= LOGLV_DEBUG {
		log.Printf("[DEBUG] " + format, args...)
	}
}

func (self *Logger) Trace(format string, args ...interface{}) {
	if self.loglv >= LOGLV_TRACE {
		log.Printf("[TRACE] " + format, args...)
	}
}

func (self *Logger) Info(format string, args ...interface{}) {
	if self.loglv >= LOGLV_INFO {
		log.Printf("[INFO] " + format, args...)
	}
}

func (self *Logger) Warn(format string, args ...interface{}) {
	if self.loglv >= LOGLV_WARN {
		log.Printf("[WARN] " + format, args...)
	}
}

func (self *Logger) Error(format string, args ...interface{}) {
	if self.loglv >= LOGLV_ERROR {
		log.Printf("[ERROR] " + format, args...)
	}
}

func (self *Logger) Fatal(format string, args ...interface{}) {
	log.Fatalf("[FATAL] " + format, args...)
}

func (self *Logger) PerformStart(key string) int64 {
	if !self.enablePerform {
		return 0
	}
	if self.performByKey[key] == nil {
		self.performByKey[key] = &PerformContext{}
	}
	return time.Now().UnixNano()
}

func (self *Logger) PerformEnd(key string, startAt int64) {
	if !self.enablePerform {
		return
	}
	if self.performByKey[key] == nil || startAt == 0 {
		self.Warn("Must call PerformStart() %s", key)
		return
	}
	self.performByKey[key].elapsed += time.Now().UnixNano() - startAt
	self.performByKey[key].count++
}

func (self *Logger) PerformDump(specifiedKey string) {
	if !self.enablePerform {
		return
	}
	if specifiedKey == "" {
		for key, perform := range(self.performByKey) {
			log.Printf("[PERFORM] %s calls: %v, elapsed: %v ms", key, perform.count, perform.elapsed / 1000000)
		}
	}
}

func InitLogger(logfile string) {
	file, err := os.OpenFile(logfile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Failure to open logfile %s", logfile)
	}
	if config.Process.Daemon {
		log.SetOutput(file)
	} else {
		log.SetOutput(io.MultiWriter(file, os.Stdout))
	}
	faissdb.logger = &Logger{loglv: LOGLV_INFO, performByKey: map[string]*PerformContext{}, enablePerform: false}
}
