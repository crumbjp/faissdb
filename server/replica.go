package main

import (
	"fmt"
	"net"
	"time"
	"errors"
	"log"
	"encoding/json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/keepalive"
	pb "github.com/crumbjp/faissdb/server/grpc_replica"
	"context"
)

const (
	FULLSYNC_BULKSIZE = 1000
	OPLOG_BULKSIZE = 1000
)

const (
	ROLE_PRIMARY = 1
	ROLE_SECONDARY = 2
)

type ReplicaSetMember struct {
	Id int `json:"id"`
	Host string `json:"host"`
	Primary bool `json:"primary"`
}

type ReplicaSet struct {
	Replica string `json:"replica"`
	Members []ReplicaSetMember `json:"members"`
}

type ReplicaMember struct {
	Id int
	Host string
	Role int
	Uuid string
	rpcClientConnection grpc.ClientConnInterface
	rpcReplicaClient pb.ReplicaClient
}

func (self *ReplicaMember) Connect() {
	clientConn, ok := self.rpcClientConnection.(*grpc.ClientConn)
	if ok {
		state := clientConn.GetState()
		if state == connectivity.Ready {
			return
		}
		faissdb.logger.Error("ReplicaMember.Connect() State is not ready %s", state)
		return
	}
	faissdb.logger.Info("ReplicaMember.Connect() New connection to %v => %v", self.Id, self.Host)
	var err error
	self.rpcClientConnection, err = grpc.Dial(
		self.Host,
		grpc.WithMaxMsgSize(100*1024*1024),
		grpc.WithInsecure())
	// self.rpcClientConnection, err = grpc.Dial(
	// 	self.Host,
	// 	grpc.WithMaxMsgSize(100*1024*1024),
	// 	grpc.WithInsecure(),
	// 	grpc.WithBlock())
	if err != nil {
		faissdb.logger.Error("ReplicaMember.Connect() grpc.Dial() %v", err)
	}
	self.rpcReplicaClient = pb.NewReplicaClient(self.rpcClientConnection)
}

func (self *ReplicaMember) Close() {
	faissdb.logger.Info("ReplicaMember.Close() %v => %v", self.Id, self.Host)
	clientConn, ok := self.rpcClientConnection.(*grpc.ClientConn)
	if ok {
		clientConn.Close()
		self.rpcClientConnection = nil
		self.rpcReplicaClient = nil
	}
}

func (self *ReplicaMember) beforeRequest(opName string, timeout time.Duration) (context.Context, context.CancelFunc, error) {
	if self.rpcReplicaClient == nil {
		return nil, nil, errors.New(fmt.Sprintf("ReplicaMember.beforeRequest() %s no client %v => %v", opName, self.Id, self.Host))
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return ctx, cancel, nil
}

func (self *ReplicaMember) GetStatus() (*pb.GetStatusReply, error){
	ctx, cancel, err := self.beforeRequest("GetStatus()", 1 * time.Second)
	if err != nil {
		faissdb.logger.Error("ReplicaMember.GetStatus() beforeRequest() %v", err)
		return nil, err
	}
	defer cancel()
	reply, err := self.rpcReplicaClient.GetStatus(ctx, &pb.GetStatusRequest{Rsts: faissdb.rsTs})
	if err != nil {
		faissdb.logger.Error("ReplicaMember.GetStatus() %v", err)
		return nil, err
	}
	return reply, nil
}

func (self *ReplicaMember) PrepareResetReplicaSet() (*pb.PrepareResetReplicaSetReply, error){
	ctx, cancel, err := self.beforeRequest("PrepareResetReplicaSet()", 5 * time.Second)
	if err != nil {
		faissdb.logger.Error("ReplicaMember.PrepareResetReplicaSet() beforeRequest() %v", err)
		return nil, err
	}
	defer cancel()
	reply, err := self.rpcReplicaClient.PrepareResetReplicaSet(ctx, &pb.PrepareResetReplicaSetRequest{})
	if err != nil {
		faissdb.logger.Error("ReplicaMember.PrepareResetReplicaSet() %v", err)
		return nil, err
	}
	return reply, nil
}

func (self *ReplicaMember) ResetReplicaSet() (*pb.ResetReplicaSetReply, error){
	ctx, cancel, err := self.beforeRequest("ResetReplicaSet()", 5 * time.Second)
	if err != nil {
		faissdb.logger.Error("ReplicaMember.ResetReplicaSet() beforeRequest() %v", err)
		return nil, err
	}
	defer cancel()
	reply, err := self.rpcReplicaClient.ResetReplicaSet(ctx, &pb.ResetReplicaSetRequest{Rsts: faissdb.rsTs, Rsjson: faissdb.rsJson})
	if err != nil {
		faissdb.logger.Error("ReplicaMember.ResetReplicaSet() %v", err)
		return nil, err
	}
	return reply, nil
}

func (self *ReplicaMember) GetCurrentOplog(startKey string, length int32) (*pb.GetCurrentOplogReply, error){
	ctx, cancel, err := self.beforeRequest("GetCurrentOplog()", 10 * time.Second)
	if err != nil {
		faissdb.logger.Error("ReplicaMember.GetCurrentOplog() beforeRequest() %v", err)
		return nil, err
	}
	defer cancel()
	reply, err := self.rpcReplicaClient.GetCurrentOplog(ctx, &pb.GetCurrentOplogRequest{Startkey: startKey, Length: length})
	if err != nil {
		faissdb.logger.Error("ReplicaMember.GetCurrentOplog() %v", err)
		return nil, err
	}
	return reply, nil
}

func (self *ReplicaMember) GetData(startKey string, length int32) (*pb.GetDataReply, error){
	ctx, cancel, err := self.beforeRequest("GetData()", 60 * time.Second)
	if err != nil {
		faissdb.logger.Error("ReplicaMember.GetData() beforeRequest() %v", err)
		return nil, err
	}
	defer cancel()
	reply, err := self.rpcReplicaClient.GetData(ctx, &pb.GetDataRequest{Startkey: startKey, Length: length})
	if err != nil {
		faissdb.logger.Error("ReplicaMember.GetData() %v", err)
		return nil, err
	}
	return reply, nil
}

func (self *ReplicaMember) GetTrained() (*pb.GetTrainedReply, error){
	ctx, cancel, err := self.beforeRequest("GetTrained()", 60 * time.Second)
	if err != nil {
		faissdb.logger.Error("ReplicaMember.GetTrained() beforeRequest() %v", err)
		return nil, err
	}
	defer cancel()
	reply, err := self.rpcReplicaClient.GetTrained(ctx, &pb.GetTrainedRequest{})
	if err != nil {
		faissdb.logger.Error("ReplicaMember.GetTrained() %v", err)
		return nil, err
	}
	return reply, nil
}

func (self *ReplicaMember) GetLastKey() (*pb.GetLastKeyReply, error){
	ctx, cancel, err := self.beforeRequest("GetLastKey()", 1 * time.Second)
	if err != nil {
		faissdb.logger.Error("ReplicaMember.GetLastKey() beforeRequest() %v", err)
		return nil, err
	}
	defer cancel()
	reply, err := self.rpcReplicaClient.GetLastKey(ctx, &pb.GetLastKeyRequest{})
	if err != nil {
		faissdb.logger.Error("ReplicaMember.GetLastKey() %v", err)
		return nil, err
	}
	return reply, nil
}

type RpcReplicaServer struct {
	pb.UnimplementedReplicaServer
}

func (self *RpcReplicaServer) PrepareResetReplicaSet(ctx context.Context, in *pb.PrepareResetReplicaSetRequest) (*pb.PrepareResetReplicaSetReply, error) {
	err := setStatus(STATUS_CONFIGURING)
	if err == nil {
		err = ReplicaSync()
	}
	return &pb.PrepareResetReplicaSetReply{Status: int32(faissdb.status), Lastkey: LastKey()}, err
}

func (self *RpcReplicaServer) ResetReplicaSet(ctx context.Context, in *pb.ResetReplicaSetRequest) (*pb.ResetReplicaSetReply, error) {
	err := ResetReplicaSet(in.GetRsts(), []byte(in.GetRsjson()))
	if err != nil {
		log.Println(err)
	}
	return &pb.ResetReplicaSetReply{Rsts: faissdb.rsTs}, err
}

func (self *RpcReplicaServer) GetStatus(ctx context.Context, in *pb.GetStatusRequest) (*pb.GetStatusReply, error) {
	rsJson := ""
	if IsPrimary() && faissdb.rsTs != in.GetRsts() {
		rsJson = faissdb.rsJson
	}
	return &pb.GetStatusReply{Rsts: faissdb.rsTs, Uuid: faissdb.selfUuid, Rsjson: rsJson, Status: int32(faissdb.status), Lastkey: LastKey()}, nil
}

func (self *RpcReplicaServer) GetLastKey(ctx context.Context, in *pb.GetLastKeyRequest) (*pb.GetLastKeyReply, error) {
	return &pb.GetLastKeyReply{Lastkey: LastKey()}, nil
}

func (self *RpcReplicaServer) GetTrained(ctx context.Context, in *pb.GetTrainedRequest) (*pb.GetTrainedReply, error) {
	data, err := ReadFaissTrained()
	return &pb.GetTrainedReply{Data: data}, err
}

func (self *RpcReplicaServer) GetData(ctx context.Context, in *pb.GetDataRequest) (*pb.GetDataReply, error) {
	keys, values, nextKey := faissdb.dataDB.GetRawData(in.GetStartkey(), int(in.GetLength()))
	return &pb.GetDataReply{Nextkey: nextKey, Keys: keys, Values: values}, nil
}

func (self *RpcReplicaServer) GetCurrentOplog(ctx context.Context, in *pb.GetCurrentOplogRequest) (*pb.GetCurrentOplogReply, error) {
	logkeys, values, err := GetCurrentOplog(in.GetStartkey(), int(in.GetLength()))
	if err != nil {
    faissdb.logger.Error("RpcReplicaServer.GetCurrentOplog() %v", err)
		return nil, err
	}
	return &pb.GetCurrentOplogReply{Keys: logkeys, Values: values}, nil
}

func IsPrimary() bool {
	return faissdb.selfMember != nil && faissdb.selfMember.Role == ROLE_PRIMARY
}

func IsSecondary() bool {
	return faissdb.selfMember != nil && faissdb.selfMember.Role == ROLE_SECONDARY
}

func PrepareResetReplicaSet() error {
	if err := setStatus(STATUS_CONFIGURING); err != nil {
		return err
	}
	for _, replicaMember := range(faissdb.replicaMembers) {
		prepareResetReplicaSetReply, err := replicaMember.PrepareResetReplicaSet()
		if err != nil {
			rollbackStatus()
			return err
		}
		if int(prepareResetReplicaSetReply.GetStatus()) != STATUS_CONFIGURING {
			return errors.New(fmt.Sprintf("PrepareResetReplicaSet() unexpected status %v", prepareResetReplicaSetReply.GetStatus()))
		}
		if prepareResetReplicaSetReply.GetLastkey() != LastKey() {
			return errors.New(fmt.Sprintf("PrepareResetReplicaSet() unexpected lastkey %v != %v", prepareResetReplicaSetReply.GetLastkey(), LastKey()))
		}
	}
	return nil
}

func ResetReplicaSet(rsts int64, jsonBytes []byte) error {
	var newReplicaSet ReplicaSet
	if err := json.Unmarshal(jsonBytes, &newReplicaSet); err != nil {
		return err
	}
	if faissdb.rsTs < rsts {
		faissdb.metaDB.PutInt64("ReplicaSetTs", rsts)
		faissdb.metaDB.PutString("ReplicaSet", string(jsonBytes))
		InitReplicaSet()
	}
	return nil
}

func checkReplicaSet(force bool) bool {
	if faissdb.rsTs == 0 {
		return false
	}
	if !force && time.Now().UnixNano() < faissdb.lastCheckedAt.Add(CHECK_REPLICASET_INTERVAL * time.Millisecond).UnixNano() {
		return (faissdb.selfMember != nil)
	}
	newSecondaryMembers := []*ReplicaMember{}
	for _, replicaMember := range(faissdb.replicaMembers) {
		getStatusReply, err := replicaMember.GetStatus()
		if err != nil {
			continue
		}
		replicaMember.Uuid = getStatusReply.GetUuid()
		if replicaMember.Uuid == faissdb.selfUuid {
			faissdb.selfMember = replicaMember
		} else if faissdb.rsTs < getStatusReply.GetRsts() {
			if err := ResetReplicaSet(getStatusReply.GetRsts(), []byte(getStatusReply.GetRsjson())); err != nil {
				log.Println(err)
			}
			return false
		} else if faissdb.rsTs > getStatusReply.GetRsts() {
			resetReplicaSetReply, err := replicaMember.ResetReplicaSet()
			if err != nil {
				faissdb.logger.Error("checkReplicaSet() replicaMember.ResetReplicaSet() %v", err)
				continue
			}
			if resetReplicaSetReply.GetRsts() != faissdb.rsTs {
				faissdb.logger.Error("checkReplicaSet() Unmatch %v %v", resetReplicaSetReply.GetRsts(), faissdb.rsTs)
				continue
			}
			getStatusReply, err = replicaMember.GetStatus()
			if err != nil {
				continue
			}
		}
		if replicaMember.Role == ROLE_PRIMARY {
			faissdb.primaryMember = replicaMember
		} else if replicaMember.Role == ROLE_SECONDARY {
			newSecondaryMembers = append(newSecondaryMembers, replicaMember)
		}
		faissdb.logger.Info("checkReplicaSet() RS: id: %v, host: %v, role: %v, self: %v", replicaMember.Id, replicaMember.Host, replicaMember.Role, replicaMember.Uuid == faissdb.selfUuid)
	}
	faissdb.secondaryMembers = newSecondaryMembers
	isValid := (faissdb.selfMember != nil)
	if isValid {
		faissdb.lastCheckedAt = time.Now()
	} else {
		faissdb.lastCheckedAt = time.Unix(0, 0)
	}
	return isValid
}

func InitReplicaSet() {
	rsts := faissdb.metaDB.GetInt64("ReplicaSetTs")
	rsJson := faissdb.metaDB.GetString("ReplicaSet")
	if rsJson == "" {
		return
	}
	var newReplicaSet ReplicaSet
	if err := json.Unmarshal([]byte(rsJson), &newReplicaSet); err != nil {
		panic(err)
	}
	if rsts != nil {
		faissdb.rsTs = *rsts
	}
	faissdb.rsJson = rsJson
	faissdb.replicaSet = &newReplicaSet
	for _, replicaMember := range(faissdb.replicaMembers) {
		replicaMember.Close()
	}
	faissdb.replicaMembers = []*ReplicaMember{}
	faissdb.selfMember = nil
	faissdb.primaryMember = nil
	faissdb.secondaryMembers = []*ReplicaMember{}
	faissdb.replicaMembers = make([]*ReplicaMember, len(faissdb.replicaSet.Members))
	for i, member := range(faissdb.replicaSet.Members) {
		faissdb.replicaMembers[i] = &ReplicaMember{Id: member.Id, Host: member.Host}
		faissdb.replicaMembers[i].Connect()
		if member.Primary {
			faissdb.replicaMembers[i].Role = ROLE_PRIMARY
		} else {
			faissdb.replicaMembers[i].Role = ROLE_SECONDARY
		}
	}
	checkReplicaSet(true)
}

func InitRpcReplicaServer() {
	faissdb.logger.Info("InitRpcReplicaServer() %v", config.Replica.Listen)
	listen, err := net.Listen("tcp", config.Replica.Listen)
	if err != nil {
    log.Fatalf("InitRpcReplicaServer() %v", err)
	}
	server := grpc.NewServer(
		grpc.MaxSendMsgSize(100*1024*1024),
		grpc.MaxRecvMsgSize(100*1024*1024),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime: 2 * time.Second,
			PermitWithoutStream: true,
		}))
	pb.RegisterReplicaServer(server, &RpcReplicaServer{})
	if err := server.Serve(listen); err != nil {
    log.Fatalf("InitRpcReplicaServer() %v", err)
	}
}

func RpcReplicaGetLastKey() (string, error) {
	if faissdb.primaryMember == nil {
		return "", errors.New("RpcReplicaGetLastKey() No primary")
	}
	reply, err := faissdb.primaryMember.GetLastKey()
	if err != nil {
		faissdb.logger.Error("RpcReplicaGetLastKey() %v", err)
		return "", err
	}
	return reply.GetLastkey(), nil
}

func RpcReplicaGetTrained() (*pb.GetTrainedReply, error) {
	if faissdb.primaryMember == nil {
		return nil, errors.New("RpcReplicaGetTrained() No primary")
	}
	reply, err := faissdb.primaryMember.GetTrained()
	if err != nil {
		faissdb.logger.Error("RpcReplicaGetTrained() %v", err)
		return nil, err
	}
	return reply, nil
}

func RpcReplicaGetData(startKey string, length int32) (*pb.GetDataReply, error) {
	if faissdb.primaryMember == nil {
		return nil, errors.New("RpcReplicaGetData() No primary")
	}
	reply, err := faissdb.primaryMember.GetData(startKey, length)
	if err != nil {
		faissdb.logger.Error("RpcReplicaGetData() %v", err)
		return nil, err
	}
	return reply, nil
}

func RpcReplicaGetCurrentOplog(startKey string, length int32) (*pb.GetCurrentOplogReply, error) {
	if faissdb.primaryMember == nil {
		return nil, errors.New("RpcReplicaGetCurrentOplog() No primary")
	}
	reply, err := faissdb.primaryMember.GetCurrentOplog(startKey, length)
	if err != nil {
		faissdb.logger.Error("RpcReplicaGetCurrentOplog() %v", err)
		return nil, err
	}
	return reply, nil
}

func ReplicaFullSync() {
	faissdb.logger.Info("ReplicaFullSync() start")
	defer faissdb.logger.Info("ReplicaFullSync() end")
	var data []byte
	for ;; {
		reply, err := RpcReplicaGetTrained()
		if err != nil {
			faissdb.logger.Error("ReplicaFullSync() RpcReplicaGetTrained() %v", err)
			time.Sleep(10 * time.Second)
			continue
		}
		data = reply.GetData()
		break
	}
	err := WriteFile(TrainedFilePath(), data)
	if err != nil {
    log.Fatalf("ReplicaFullSync() %v", err)
	}
	localIndex.ResetToTrained()
	var masterLastKey string
	masterLastKey, err = RpcReplicaGetLastKey()
	faissdb.logger.Info("ReplicaFullSync() masterLastKey: %s", masterLastKey)
	currentKey := ""
	for ;; {
		reply, err := RpcReplicaGetData(currentKey, FULLSYNC_BULKSIZE)
		if err != nil {
			log.Fatalf("ReplicaFullSync() %v", err)
		}
		for i, value := range reply.GetValues() {
			faissdbRecord := &pb.FaissdbRecord{}
			DecodeFaissdbRecord(faissdbRecord, value)
			SetRaw(reply.GetKeys()[i], faissdbRecord)
		}
		if reply.GetNextkey() == "" {
			break
		}
		currentKey = reply.GetNextkey()
		faissdb.logger.Info("ReplicaFullSync() next: %s", currentKey)
	}
	PutOplogWithKey(masterLastKey, OP_SYSTEM, "", []byte("FullSync"))
	ReplicaSync()
}

func ApplyOplog(oplog *Oplog) error {
	if oplog.op == OP_SET {
		faissdbRecord := &pb.FaissdbRecord{}
		performDecodeFaissdbRecord := faissdb.logger.PerformStart("ApplyOplog DecodeFaissdbRecord")
		err := DecodeFaissdbRecord(faissdbRecord, oplog.d)
		faissdb.logger.PerformEnd("ApplyOplog DecodeFaissdbRecord", performDecodeFaissdbRecord)
		if err != nil {
			faissdb.logger.Error("ApplyOplog() DecodeFaissdbRecord() %v", err)
			return err
		}
		performSetRaw := faissdb.logger.PerformStart("ApplyOplog SetRaw")
		SetRaw(oplog.key, faissdbRecord)
		faissdb.logger.PerformEnd("ApplyOplog SetRaw", performSetRaw)
	} else if oplog.op == OP_DEL {
		faissdbRecord := &pb.FaissdbRecord{}
		performDecodeFaissdbRecord := faissdb.logger.PerformStart("ApplyOplog DecodeFaissdbRecord del")
		err := DecodeFaissdbRecord(faissdbRecord, oplog.d)
		faissdb.logger.PerformEnd("ApplyOplog DecodeFaissdbRecord del", performDecodeFaissdbRecord)
		if err != nil {
			faissdb.logger.Error("ApplyOplog() DecodeFaissdbRecord() %v", err)
			return err
		}
		performDelRaw := faissdb.logger.PerformStart("ApplyOplog DelRaw")
		DelRaw(oplog.key, faissdbRecord)
		faissdb.logger.PerformEnd("ApplyOplog DelRaw", performDelRaw)
	} else if oplog.op == OP_SYSTEM {
	}
	return nil
}

func ReplicaSync() error {
	if IsPrimary() {
		return nil
	}
	faissdb.replicaSyncMutex.Lock()
	defer faissdb.replicaSyncMutex.Unlock()
	for ;; {
		performRpcReplicaGetCurrentOplog := faissdb.logger.PerformStart("ReplicaSync RpcReplicaGetCurrentOplog")
		lastKey := LastKey()
		reply, err := RpcReplicaGetCurrentOplog(lastKey, OPLOG_BULKSIZE)
		if err != nil {
			faissdb.logger.Error("ReplicaSync() RpcReplicaGetCurrentOplog() %v", err)
			return err
		}
		faissdb.logger.PerformEnd("ReplicaSync RpcReplicaGetCurrentOplog", performRpcReplicaGetCurrentOplog)
		for i, value := range reply.GetValues() {
			oplog := &Oplog{}
			performDecodeOplog := faissdb.logger.PerformStart("ReplicaSync Decode")
			oplog.Decode(value)
			faissdb.logger.PerformEnd("ReplicaSync Decode", performDecodeOplog)
			performApplyOplog := faissdb.logger.PerformStart("ReplicaSync Apply")
			err = ApplyOplog(oplog)
			faissdb.logger.PerformEnd("ReplicaSync Apply", performApplyOplog)
			if err != nil {
				return err
			}
			performPutOplog := faissdb.logger.PerformStart("ReplicaSync PutOplogWithKey")
			PutOplogWithKey(reply.GetKeys()[i], oplog.op, oplog.key, oplog.d)
			faissdb.logger.PerformEnd("ReplicaSync PutOplogWithKey", performPutOplog)
		}
		faissdb.logger.PerformDump("")
		valueSize := len(reply.GetValues())
		if valueSize > 0 {
			now, idx := faissdb.oplogKeyGenerator.Parse(lastKey)
			faissdb.logger.Trace("ReplicaSync() lastkey: %v (%v %v) n: %v", lastKey, now.Format("2006-01-02 15:04:05.000"), idx, valueSize)
		}
		if valueSize != OPLOG_BULKSIZE {
			break
		}
	}
	return nil
}

func InitReplicaSyncThread() {
	for ;; {
		time.Sleep(1000 * time.Millisecond)
		if !checkReplicaSet(false) {
			faissdb.logger.Warn("InitReplicaSyncThread() invalid")
			continue
		}
		if !faissdb.firstSync {
			faissdb.firstSync = true
			if IsSecondary() && LastKey() == "" {
				ReplicaFullSync()
			}
		}
		if ReplicaSync() == nil {
			setStatus(STATUS_READY)
		}
	}
}
