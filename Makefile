export GOBIN := $(shell pwd)/bin

PB_REL="https://github.com/protocolbuffers/protobuf/releases"

tools: install_protoc
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/vektra/mockery/v2@v2.43.2


install_protoc:
	curl -LO $(PB_REL)/download/v27.3/protoc-27.3-osx-x86_64.zip && \
	unzip protoc-27.3-osx-x86_64.zip -d ./bin && \
	rm protoc-27.3-osx-x86_64.zip

generate:
	go generate ./...
	go mod tidy


run:
	go run -ldflags "-X 'main.buildVersion=v1.0.1' -X 'main.buildDate=$(date)' -X 'main.buildCommit=blank'" cmd/shortener/main.go \
	-a :7070 \
	-ga :3200 \
	-b "http://localhost:7070" \
	-d "postgresql://admin:admin@localhost:5432/url_shortener?sslmode=disable" \
	-s \
	-c "config.json"

lint:
	go vet ./...

#go vet ./foo ./internal/...

custom_lint:
	go build -o ./bin/staticlint ./cmd/staticlint/main.go && ./bin/staticlint ./...

install_doc:
	go install -v golang.org/x/tools/cmd/godoc@latest

doc:
	godoc -http=:8088