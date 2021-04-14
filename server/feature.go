package main

import (
	"net"
	"time"
	"log"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	pb "github.com/crumbjp/faissdb/server/grpc_feature"
	"context"
)

type RpcFeatureServer struct {
	pb.UnimplementedFeatureServer
}

func (self *RpcFeatureServer) Set(ctx context.Context, in *pb.SetRequest) (*pb.SetReply, error) {
	nStored := 0
	nError := 0
	if IsTraining() || terminating {
		return nil, errors.New("Not ready")
	}
	var err error
	for _, data := range in.GetData() {
		v := make([]float32, config.Db.Faiss.Dimension)
		if(data.GetV() != nil) {
			for i, double := range data.GetV() {
				v[i] = float32(double)
			}
		} else {
			err := parseSparseV(v, data.GetSparsev())
			if err != nil {
				log.Println(err)
				nError++
				continue
			}
		}
		log.Println(" - set data", data.GetKey(), len(v))
		err = Set(data.GetKey(), v)
		if err != nil {
			log.Println(err)
			nError++
		} else {
			nStored++
		}
	}
	return &pb.SetReply{Nstored: int32(nStored), Nerror: int32(nError)}, nil
}

func (self *RpcFeatureServer) Del(ctx context.Context, in *pb.DelRequest) (*pb.DelReply, error) {
	if IsTraining() || terminating {
		return nil, errors.New("Not ready")
	}
	for _, key := range in.GetKey() {
		log.Println(" - del key", key)
		Del(key)
	}
	return &pb.DelReply{}, nil
}


func (self *RpcFeatureServer) Train(ctx context.Context, in *pb.TrainRequest) (*pb.TrainReply, error) {
	if IsTraining() || terminating {
		return nil, errors.New("Not ready")
	}
	log.Println(" - train")
	err := Train(float32(in.GetProportion()), in.GetForce())
	return &pb.TrainReply{}, err
}

func (self *RpcFeatureServer) Search(ctx context.Context, in *pb.SearchRequest) (*pb.SearchReply, error) {
	v := make([]float32, config.Db.Faiss.Dimension)
	if(in.GetV() != nil) {
		for i, double := range in.GetV() {
			v[i] = float32(double)
		}
	} else {
		err := parseSparseV(v, in.GetSparsev())
		if err != nil {
			return nil, err
		}
	}
	log.Println(" - search", v, in.GetN())
	searchResults := Search(v, int64(in.GetN()))
	keys := make([]string, len(searchResults))
	distances := make([]float64, len(searchResults))
	for i, searchResult := range searchResults {
		keys[i] = searchResult.key
		distances[i] = float64(searchResult.distance)
	}
	return &pb.SearchReply{Distances: distances, Keys: keys}, nil
}

func InitRpcFeatureServer() {
	listen, err := net.Listen("tcp", config.Feature.Listen)
	if err != nil {
    log.Fatalf("InitRpcFeatureServer() %v", err)
	}
	server := grpc.NewServer(
		grpc.MaxSendMsgSize(100*1024*1024),
		grpc.MaxRecvMsgSize(100*1024*1024),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime: 2 * time.Second,
			PermitWithoutStream: true,
		}))
	pb.RegisterFeatureServer(server, &RpcFeatureServer{})
	if err := server.Serve(listen); err != nil {
    log.Fatalf("InitRpcFeatureServer() %v", err)
	}
}


// var rpcClientConnection grpc.ClientConnInterface
// var rpcFeatureClient pb.FeatureClient

// func rpcFeatureConnect() {
// 	clientConn, ok := rpcClientConnection.(*grpc.ClientConn)
// 	if ok {
// 		state := clientConn.GetState()
// 		if state == connectivity.Ready {
// 			return
// 		}
// 		log.Printf("Connection error %s", state)
// 		// clientConn.Close()
// 		// rpcClientConnection = nil
// 		return
// 	}
// 	log.Printf("New connection")
// 	var err error
// 	rpcClientConnection, err = grpc.Dial(config.Feature.Master, grpc.WithInsecure(), grpc.WithBlock())
// 	if err != nil {
// 		log.Printf("InitRpcClient %v", err)
// 	}
// 	log.Printf("connected")
// 	rpcFeatureClient = pb.NewFeatureClient(rpcClientConnection)
// }

// func InitRpcFeatureClient() {
// 	if IsMaster() {
// 		return
// 	}
// 	rpcFeatureConnect()
// }

// func RpcFeatureSet() (string, error) {
// 	if rpcFeatureClient == nil {
// 		return "", errors.New("RpcFeatureSet() no client")
// 	}
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 	defer cancel()
// 	reply, err := rpcFeatureClient.Set(ctx, &pb.SetRequest{})
// 	if err != nil {
// 		log.Printf("RpcFeatureSet %v", err)
// 		return "", err
// 	}
// 	return reply.GetLastkey(), nil
// }

// func RpcFeatureDel() ([]byte, error) {
// 	if rpcFeatureClient == nil {
// 		return nil, errors.New("RpcFeatureDel() no client")
// 	}
// 	ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
// 	defer cancel()
// 	reply, err := rpcFeatureClient.Del(ctx, &pb.DelRequest{})
// 	if err != nil {
// 		log.Printf("RpcFeatureDel %v", err)
// 		return nil, err
// 	}
// 	return reply.Search(), nil
// }

// func RpcFeatureSearch(startKey string, length int32) ([]string, [][]byte, string, error) {
// 	if rpcFeatureClient == nil {
// 		return nil, nil, "", errors.New("RpcFeatureSearch() no client")
// 	}
// 	ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
// 	defer cancel()
// 	reply, err := rpcFeatureClient.Search(ctx, &pb.SearchRequest{Startkey: startKey, Length: length})
// 	if err != nil {
// 		log.Printf("RpcFeatureSearch %v", err)
// 		return nil, nil, "", err
// 	}
// 	return reply.GetKeys(), reply.GetValues(), reply.GetNextkey(), nil
// }

// func RpcFeatureTrain(startKey string, length int32) ([]string, [][]byte, error) {
// 	if rpcFeatureClient == nil {
// 		return nil, nil, errors.New("RpcFeatureTrain() no client")
// 	}
// 	ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
// 	defer cancel()
// 	reply, err := rpcFeatureClient.Train(ctx, &pb.TrainRequest{Startkey: startKey, Length: length})
// 	if err != nil {
// 		log.Printf("RpcFeatureTrain %v", err)
// 		return nil, nil, err
// 	}
// 	return reply.GetKeys(), reply.GetValues(), nil
// }
