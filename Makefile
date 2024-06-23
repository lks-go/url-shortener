generate:
	go generate ./...

run:
	go run cmd/shortener/main.go \
	-a :7070 \
	-b "http://localhost:7070" \
	-d "postgresql://admin:admin@localhost:5432/url_shortener?sslmode=disable"

