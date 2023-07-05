# START: begin
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

# END: begin

# START: client
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=certificates/ca-config.json \
		-profile=client \
		certificates/client-csr.json | cfssljson -bare client
# END: client

# START: multi_client
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=certificates/ca-config.json \
		-profile=client \
		-cn="root" \
		certificates/client-csr.json | cfssljson -bare root-client

	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=certificates/ca-config.json \
		-profile=client \
		-cn="nobody" \
		certificates/client-csr.json | cfssljson -bare nobody-client
# END: multi_client

# START: begin
	mv *.pem *.csr ${CONFIG_PATH}

# END: begin
# START: auth
$(CONFIG_PATH)/access-control-model.conf:
	cp certificates/access-control-model.conf $(CONFIG_PATH)/access-control-model.conf

$(CONFIG_PATH)/access-control-policy.csv:
	cp certificates/access-control-policy.csv $(CONFIG_PATH)/access-control-policy.csv

# START: begin
.PHONY: test
# END: auth
test:
# END: begin
# START: auth
test: $(CONFIG_PATH)/access-control-policy.csv $(CONFIG_PATH)/access-control-model.conf
#: START: begin
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...
# END: auth

.PHONY: build
build:
	protoc proto/v1/*.proto --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --proto_path=.
# END: begin