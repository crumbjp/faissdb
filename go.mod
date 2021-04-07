module server.go

go 1.15

replace local.packages/go-faiss => ./go-faiss

require (
	github.com/tecbot/gorocksdb v0.0.0-20191217155057-f0fad39f321c
	golang.org/x/net v0.0.0-20210331212208-0fccb6fa2b5c
	gopkg.in/yaml.v2 v2.4.0
	local.packages/go-faiss v0.0.0-00010101000000-000000000000
)
