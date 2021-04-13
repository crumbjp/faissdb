all: faissdb

faissdb: *.go rpcreplica/rpcreplica.pb.go
	go build

rpcreplica/rpcreplica.pb.go: rpcreplica/rpcreplica.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative rpcreplica/rpcreplica.proto
