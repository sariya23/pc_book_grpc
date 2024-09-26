gen_pb:
	protoc -I ./proto proto/*.proto  --go_out=./pb/ --go_opt=paths=source_relative --go-grpc_out=./pb/ --go-grpc_opt=paths=source_relative

clean_pb:
	rm pb/*.go

server:
	go run cmd/server/main.go -port 8080

client:
	go run cmd/client/main.go -addr 0.0.0.0:8080

test:
	go test -cover -race ./...

.PHONY: gen_pb clean_pb server test client