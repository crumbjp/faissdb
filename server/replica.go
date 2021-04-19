package main

import (
	"net"
	"time"
	"errors"
	"log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/keepalive"
	pb "github.com/crumbjp/faissdb/server/grpc_replica"
	"context"
)

type RpcReplicaServer struct {
	pb.UnimplementedReplicaServer
}

func (self *RpcReplicaServer) GetLastKey(ctx context.Context, in *pb.GetLastKeyRequest) (*pb.GetLastKeyReply, error) {
	return &pb.GetLastKeyReply{Lastkey: LastKey()}, nil
}

func (self *RpcReplicaServer) GetTrained(ctx context.Context, in *pb.GetTrainedRequest) (*pb.GetTrainedReply, error) {
	data, err := ReadFaissTrained()
	return &pb.GetTrainedReply{Data: data}, err
}

func (self *RpcReplicaServer) GetData(ctx context.Context, in *pb.GetDataRequest) (*pb.GetDataReply, error) {
	keys, slices, nextKey := GetRawData(in.GetStartkey(), int(in.GetLength()))
	values := make([][]byte, len(slices))
	for i, slice := range slices {
		values[i] = slice.Data()
		defer slice.Free()
	}
	return &pb.GetDataReply{Nextkey: nextKey, Keys: keys, Values: values}, nil
}

func (self *RpcReplicaServer) GetCurrentOplog(ctx context.Context, in *pb.GetCurrentOplogRequest) (*pb.GetCurrentOplogReply, error) {
	keys, slices, err := GetCurrentOplog(in.GetStartkey(), int(in.GetLength()))
	if err != nil {
    log.Printf("GetCurrentOplog() %v", err)
		return nil, err
	}
	values := make([][]byte, len(slices))
	for i, slice := range slices {
		values[i] = slice.Data()
		defer slice.Free()
	}
	return &pb.GetCurrentOplogReply{Keys: keys, Values: values}, nil
}

func InitRpcReplicaServer() {
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


var rpcClientConnection grpc.ClientConnInterface
var rpcReplicaClient pb.ReplicaClient

func rpcReplicaConnect() {
	clientConn, ok := rpcClientConnection.(*grpc.ClientConn)
	if ok {
		state := clientConn.GetState()
		if state == connectivity.Ready {
			return
		}
		log.Printf("Connection error %s", state)
		// clientConn.Close()
		// rpcClientConnection = nil
		return
	}
	log.Printf("New connection")
	var err error
	rpcClientConnection, err = grpc.Dial(
		config.Replica.Master,
		grpc.WithMaxMsgSize(100*1024*1024),
		grpc.WithInsecure(),
		grpc.WithBlock())
	if err != nil {
		log.Printf("InitRpcClient %v", err)
	}
	log.Printf("connected")
	rpcReplicaClient = pb.NewReplicaClient(rpcClientConnection)
}

func InitRpcReplicaClient() {
	if IsMaster() {
		return
	}
	rpcReplicaConnect()
}

func RpcReplicaGetLastKey() (string, error) {
	if rpcReplicaClient == nil {
		return "", errors.New("RpcReplicaGetLastKey() no client")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	reply, err := rpcReplicaClient.GetLastKey(ctx, &pb.GetLastKeyRequest{})
	if err != nil {
		log.Printf("RpcReplicaGetLastKey %v", err)
		return "", err
	}
	return reply.GetLastkey(), nil
}

func RpcReplicaGetTrained() ([]byte, error) {
	if rpcReplicaClient == nil {
		return nil, errors.New("RpcReplicaGetTrained() no client")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
	defer cancel()
	reply, err := rpcReplicaClient.GetTrained(ctx, &pb.GetTrainedRequest{})
	if err != nil {
		log.Printf("RpcReplicaGetTrained %v", err)
		return nil, err
	}
	return reply.GetData(), nil
}

func RpcReplicaGetData(startKey string, length int32) ([]string, [][]byte, string, error) {
	if rpcReplicaClient == nil {
		return nil, nil, "", errors.New("RpcReplicaGetData() no client")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
	defer cancel()
	reply, err := rpcReplicaClient.GetData(ctx, &pb.GetDataRequest{Startkey: startKey, Length: length})
	if err != nil {
		log.Printf("RpcReplicaGetData %v", err)
		return nil, nil, "", err
	}
	return reply.GetKeys(), reply.GetValues(), reply.GetNextkey(), nil
}

func RpcReplicaGetCurrentOplog(startKey string, length int32) ([]string, [][]byte, error) {
	if rpcReplicaClient == nil {
		return nil, nil, errors.New("RpcReplicaGetCurrentOplog() no client")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
	defer cancel()
	reply, err := rpcReplicaClient.GetCurrentOplog(ctx, &pb.GetCurrentOplogRequest{Startkey: startKey, Length: length})
	if err != nil {
		log.Printf("RpcReplicaGetCurrentOplog %v", err)
		return nil, nil, err
	}
	return reply.GetKeys(), reply.GetValues(), nil
}


func IsMaster() bool {
	return config.Replica.Master == ""
}

const (
	FULLSYNC_BULKSIZE = 10
	OPLOG_BULKSIZE = 10
)

func ReplicaFullSync() {
	log.Printf("ReplicaFullSync() start")
	var err error
	var data []byte
	for ;; {
		data, err = RpcReplicaGetTrained()
		if err != nil {
			log.Printf("ReplicaFullSync() %v", err)
			time.Sleep(1000 * time.Millisecond)
			continue
		}
		break
	}
	err = WriteFile(TrainedFilePath(), data)
	if err != nil {
    log.Fatalf("ReplicaFullSync() %v", err)
	}
	localIndex.ResetToTrained()
	var masterLastKey string
	masterLastKey, err = RpcReplicaGetLastKey()
	log.Printf("ReplicaFullSync() masterLastKey: %s", masterLastKey)
	currentKey := ""
	for ;; {
		keys, values, nextKey, err := RpcReplicaGetData(currentKey, FULLSYNC_BULKSIZE)
		if err != nil {
			log.Fatalf("ReplicaFullSync() %v", err)
		}
		for i, value := range values {
			faissdbRecord := &pb.FaissdbRecord{}
			DecodeFaissdbRecord(faissdbRecord, value)
			SetRaw(keys[i], faissdbRecord)
		}
		if nextKey == "" {
			break
		}
		currentKey = nextKey
		log.Printf("ReplicaFullSync() next: %s", currentKey)
	}
	PutOplogWithKey(masterLastKey, OP_SYSTEM, "", []byte("FullSync"))
	log.Printf("ReplicaFullSync() end")
	ReplicaSync()
}

func ReplicaSync() error {
	for ;; {
		lastKey := LastKey()
		log.Printf("ReplicaSync() lastkey: %v", lastKey)
		keys, values, err := RpcReplicaGetCurrentOplog(lastKey, OPLOG_BULKSIZE)
		if err != nil {
			log.Printf("ReplicaSync() %v", err)
			return err
		}
		for i, value := range values {
			oplog := Oplog{}
			oplog.Decode(value)
			if oplog.op == OP_SET {
				faissdbRecord := &pb.FaissdbRecord{}
				err = DecodeFaissdbRecord(faissdbRecord, oplog.d)
				if err != nil {
					log.Printf("ReplicaSync() %v", err)
					return err
				}
				SetRaw(oplog.key, faissdbRecord)
				PutOplogWithKey(keys[i], OP_SET, oplog.key, oplog.d)
			} else if oplog.op == OP_DEL {
				faissdbRecord := &pb.FaissdbRecord{}
				err = DecodeFaissdbRecord(faissdbRecord, oplog.d)
				if err != nil {
					log.Printf("ReplicaSync() %v", err)
					return err
				}
				DelRaw(keys[i], faissdbRecord)
				PutOplogWithKey(keys[i], OP_DEL, oplog.key, oplog.d)
			} else if oplog.op == OP_SYSTEM {
				PutOplogWithKey(keys[i], OP_SYSTEM, oplog.key, oplog.d)
			}
		}
		if len(values) != OPLOG_BULKSIZE {
			break
		}
	}
	return nil
}

func InitReplicaSyncThread() {
	for ;; {
		ReplicaSync()
		time.Sleep(1000 * time.Millisecond)
	}
}
