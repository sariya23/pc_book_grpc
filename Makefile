gen_pb:
	protoc -I ./proto proto/*.proto  --go_out=./pb/ --go_opt=paths=source_relative --go-grpc_out=./pb/ --go-grpc_opt=paths=source_relative

clean_pb:
	rm pb/*.go

server:
	go run cmd/server/main.go -port 8080

client:
	go run cmd/client/main.go

test:
	go test -cover -race ./...

