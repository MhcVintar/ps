.PHONY: default
default:
	$(MAKE) protoc
	$(MAKE) build

.PHONY: protoc
protoc:
	protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		internal/api/public.proto
	protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		internal/api/internal.proto

.PHONY: build
build:
	go build -o ./build/server ./cmd/server
	go build -o ./build/control ./cmd/control
	go build -o ./build/client ./cmd/client
	go build -o ./build/demo ./cmd/demo

.PHONY: buildClient
buildClient:
	go build -o build/client cmd/client/main.go
