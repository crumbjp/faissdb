package main

import (
	"net"
	"time"
	"errors"
	"log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	pb "github.com/crumbjp/faissdb/server/rpcreplica"
	"context"
)

type RpcServer struct {
	pb.UnimplementedReplicaServer
}

func (self *RpcServer) GetLastKey(ctx context.Context, in *pb.GetLastKeyRequest) (*pb.GetLastKeyReply, error) {
	return &pb.GetLastKeyReply{Lastkey: LastKey()}, nil
}

func (self *RpcServer) GetTrained(ctx context.Context, in *pb.GetTrainedRequest) (*pb.GetTrainedReply, error) {
	return &pb.GetTrainedReply{Data: ReadFaissTrained()}, nil
}

func (self *RpcServer) GetData(ctx context.Context, in *pb.GetDataRequest) (*pb.GetDataReply, error) {
	keys, slices, nextKey := GetRawData(in.GetStartkey(), int(in.GetLength()))
	values := make([][]byte, len(slices))
	for i, slice := range slices {
		values[i] = slice.Data()
		defer slice.Free()
	}
	return &pb.GetDataReply{Nextkey: nextKey, Keys: keys, Values: values}, nil
}

func (self *RpcServer) GetCurrentOplog(ctx context.Context, in *pb.GetCurrentOplogRequest) (*pb.GetCurrentOplogReply, error) {
	keys, slices, err := GetCurrentOplog(in.GetStartkey(), int(in.GetLength()))
	if err != nil {
    log.Println("GetCurrentOplog() %v", err)
		return nil, err
	}
	values := make([][]byte, len(slices))
	for i, slice := range slices {
		values[i] = slice.Data()
		defer slice.Free()
	}
	return &pb.GetCurrentOplogReply{Keys: keys, Values: values}, nil
}

func InitRpcServer() {
	listen, err := net.Listen("tcp", config.Replica.Listen)
	if err != nil {
    log.Fatalf("InitRpcServer() %v", err)
	}
	server := grpc.NewServer()
	pb.RegisterReplicaServer(server, &RpcServer{})
	if err := server.Serve(listen); err != nil {
    log.Fatalf("InitRpcServer() %v", err)
	}
}


var rpcClientConnection grpc.ClientConnInterface
var replicaClient pb.ReplicaClient

func RpcConnect() {
	clientConn, ok := rpcClientConnection.(*grpc.ClientConn)
	if ok {
		state := clientConn.GetState()
		if state == connectivity.Ready {
			return
		}
		log.Println("Connection error %s", state)
		clientConn.Close()
		rpcClientConnection = nil
	}
	log.Println("New connection")
	var err error
	rpcClientConnection, err = grpc.Dial(config.Replica.Master, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Printf("InitRpcClient %v", err)
	}
	log.Println("connected")
	replicaClient = pb.NewReplicaClient(rpcClientConnection)
}

func RpcKeepConnectionThread() {
	for ;; {
		RpcConnect()
		time.Sleep(1000 * time.Millisecond)
	}
}

func InitRpcClient() {
	if IsMaster() {
		return
	}
	RpcConnect()
	go RpcKeepConnectionThread()
}

func RpcGetLastKey() (string, error) {
	if replicaClient == nil {
		return "", errors.New("RpcGetLastKey() no client")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	reply, err := replicaClient.GetLastKey(ctx, &pb.GetLastKeyRequest{})
	if err != nil {
		log.Printf("InitRpcClient %v", err)
		return "", err
	}
	return reply.GetLastkey(), nil
}

func RpcGetTrained() ([]byte, error) {
	if replicaClient == nil {
		return nil, errors.New("RpcGetTrained() no client")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
	defer cancel()
	reply, err := replicaClient.GetTrained(ctx, &pb.GetTrainedRequest{})
	if err != nil {
		log.Printf("InitRpcClient %v", err)
		return nil, err
	}
	return reply.GetData(), nil
}

func RpcGetData(startKey string, length int32) ([]string, [][]byte, string, error) {
	if replicaClient == nil {
		return nil, nil, "", errors.New("RpcGetData() no client")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
	defer cancel()
	reply, err := replicaClient.GetData(ctx, &pb.GetDataRequest{Startkey: startKey, Length: length})
	if err != nil {
		log.Printf("InitRpcClient %v", err)
		return nil, nil, "", err
	}
	return reply.GetKeys(), reply.GetValues(), reply.GetNextkey(), nil
}

func RpcGetCurrentOplog(startKey string, length int32) ([]string, [][]byte, error) {
	if replicaClient == nil {
		return nil, nil, errors.New("RpcGetCurrentOplog() no client")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
	defer cancel()
	reply, err := replicaClient.GetCurrentOplog(ctx, &pb.GetCurrentOplogRequest{Startkey: startKey, Length: length})
	if err != nil {
		log.Printf("InitRpcClient %v", err)
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
	masterLastKey, err := RpcGetLastKey()
	data, err := RpcGetTrained()
	if err != nil {
    log.Fatalf("ReplicaFullSync() %v", err)
	}
	err = WriteFile(IndexFilePath(), data)
	if err != nil {
    log.Fatalf("ReplicaFullSync() %v", err)
	}
	err = WriteFile(TrainedFilePath(), data)
	if err != nil {
    log.Fatalf("ReplicaFullSync() %v", err)
	}
	InitLocalIndex()
	currentKey := ""
	for ;; {
		keys, values, nextKey, err := RpcGetData(currentKey, FULLSYNC_BULKSIZE)
		if err != nil {
			log.Fatalf("ReplicaFullSync() %v", err)
		}
		for i, value := range values {
			data := Data{}
			data.Decode(value)
			SetData(keys[i], data)
		}
		if nextKey == "" {
			break
		}
		currentKey = nextKey
	}
	PutOplogWithKey(masterLastKey, OP_SYSTEM, "", []byte("FullSync"))
	ReplicaSync()
}

func ReplicaSync() error {
	for ;; {
		lastKey := LastKey()
		keys, values, err := RpcGetCurrentOplog(lastKey, OPLOG_BULKSIZE)
		if err != nil {
			log.Println("ReplicaSync() %v", err)
			return err
		}
		for i, value := range values {
			oplog := Oplog{}
			oplog.Decode(value)
			if oplog.op == OP_SET {
				data := Data{}
				err = data.Decode(oplog.d)
				if err != nil {
					log.Println("ReplicaSync() %v", err)
					return err
				}
				encoded := SetData(oplog.key, data)
				PutOplog(OP_SET, keys[i], encoded)
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
