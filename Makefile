gen_pb:
	protoc -I ./proto proto/*.proto \
	--go_out=./pb/ --go_opt=paths=source_relative \
	--go-grpc_out=./pb/ --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=:pb --swagger_out=:swagger

clean_pb:
	rm pb/*.go

server1:
	go run cmd/server/main.go -port 50051

server2:
	go run cmd/server/main.go -port 50052

server:
	go run cmd/server/main.go -port 8080

rest:
	go run cmd/server/main.go -port 8081 -type grpc

client:
	go run cmd/client/main.go -addr 0.0.0.0:8080

test:
	go test -cover -race ./...

cert:
	cd cert; ./gen.sh; cd ..

.PHONY: gen_pb clean_pb server test client cert

