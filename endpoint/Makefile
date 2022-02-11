
all: client server

podman-build:
	podman build ./server/ -t quay.io/fennec-project/snoopy-data-endpoint:0.0.1-4

podman-push:
	podman push quay.io/fennec-project/snoopy-data-endpoint:0.0.1-4

protoc:
	@echo "Generating Go files"
	cd proto && protoc --go_out=. --go-grpc_out=. \
		--go-grpc_opt=paths=source_relative --go_opt=paths=source_relative *.proto

server: protoc
	@echo "Building server"
	go build -o server \
		github.com/fennec-project/snoopy-operator/endpoint/server/

client: protoc
	@echo "Building client"
	go build -o client \
		github.com/fennec-project/snoopy-operator/endpoint/client

clean:
	go clean github.com/fennec-project/snoopy-operator/endpoint/...
	rm -f server client

.PHONY: client server protoc