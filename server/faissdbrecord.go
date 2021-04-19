package main

import (
	pb "github.com/crumbjp/faissdb/server/grpc_replica"
	"google.golang.org/protobuf/proto"
)

func DecodeFaissdbRecord(faissdbRecord *pb.FaissdbRecord, b []byte) error {
	return proto.Unmarshal(b, faissdbRecord)
}

func EncodeFaissdbRecord(faissdbRecord *pb.FaissdbRecord) ([]byte, error) {
	return proto.Marshal(faissdbRecord)
}
