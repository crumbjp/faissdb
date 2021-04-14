all: faissdb

faissdb: *.go grpc_replica/replica.pb.go grpc_feature/feature.pb.go
	go build

grpc_replica/replica.pb.go: ../protos/replica.proto
	protoc --proto_path=../protos --go_out=./grpc_replica --go_opt=paths=source_relative --go-grpc_out=./grpc_replica --go-grpc_opt=paths=source_relative ../protos/replica.proto

grpc_feature/feature.pb.go: ../protos/feature.proto
	protoc --proto_path=../protos --go_out=./grpc_feature --go_opt=paths=source_relative --go-grpc_out=./grpc_feature --go-grpc_opt=paths=source_relative ../protos/feature.proto
