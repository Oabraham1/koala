CONFIG_PATH=${HOME}/.koala/

.PHONY: init
init:
	mkdir -p ${CONFIG_PATH}

.PHONY: generate-certificate
generate-certificate:
	cfssl gencert -initca certificates/ca-csr.json | cfssljson -bare ca

	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=certificates/ca-config.json \
		-profile=server \
		certificates/server-csr.json | cfssljson -bare server

.PHONY: test
test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: build
build:
	protoc proto/v1/*.proto --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --proto_path=.