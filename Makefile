proto-gen:
	protoc proto/v1/*.proto --go_out=. --go_opt=paths=source_relative --proto_path=.

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...