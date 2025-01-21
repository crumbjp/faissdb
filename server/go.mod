module faissdb

go 1.23.4

replace github.com/crumbjp/faissdb/server => ./

require (
	github.com/crumbjp/faissdb/server v0.0.0-20240207154259-83f6225dc708
	github.com/crumbjp/go-faiss v0.2.0
	github.com/google/uuid v1.6.0
	github.com/linxGnu/grocksdb v1.9.8
	github.com/sevlyar/go-daemon v0.1.6
	github.com/stretchr/testify v1.10.0
	golang.org/x/net v0.34.0
	google.golang.org/grpc v1.69.4
	google.golang.org/protobuf v1.36.2
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250106144421-5f5ef82da422 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
