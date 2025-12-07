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
	go build -o build/server cmd/server/main.go
	go build -o build/control cmd/control/main.go