OUTPUT=src/grpc_feature
NPM_BIN=$(shell npm bin)

GRPC_TOOL=$(NPM_BIN)/grpc_tools_node_protoc
TYPESCRIPT_PLUGIN=protoc-gen-ts=$(NPM_BIN)/protoc-gen-ts

all: src/grpc_feature/feature_grpc_pb.js

src/grpc_feature/feature_grpc_pb.js: ../protos/feature.proto
	$(GRPC_TOOL) --plugin=${TYPESCRIPT_PLUGIN} --js_out=import_style=commonjs,binary:$(OUTPUT) --grpc_out=grpc_js:$(OUTPUT) --ts_out=grpc_js:$(OUTPUT) -I ../protos/ ../protos/feature.proto
