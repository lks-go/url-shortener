export GOBIN := $(shell pwd)/bin

tools:
	go install github.com/vektra/mockery/v2@v2.43.2

generate:
	go generate ./...

run:
	go run -ldflags "-X 'main.buildVersion=v1.0.1' -X 'main.buildDate=$(date)' -X 'main.buildCommit=blank'" cmd/shortener/main.go \
	-a :7070 \
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