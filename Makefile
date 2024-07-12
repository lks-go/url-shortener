generate:
	go generate ./...

run:
	go run cmd/shortener/main.go \
	-a :7070 \
	-b "http://localhost:7070" \
	-d "postgresql://admin:admin@localhost:5432/url_shortener?sslmode=disable"

lint:
	go vet ./...

#go vet ./foo ./internal/...

custom_lint:
	go build -o ./bin/staticlint ./cmd/staticlint/main.go && ./bin/staticlint ./...

install_doc:
	go install -v golang.org/x/tools/cmd/godoc@latest

doc:
	godoc -http=:8088