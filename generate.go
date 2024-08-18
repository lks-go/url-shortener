package gobone

//go:generate ./bin/mockery --all --dir ./internal/transport/httphandlers/ 		--output ./internal/transport/httphandlers/mocks
//go:generate ./bin/mockery --all --dir ./internal/service/ 					--output ./internal/service/mocks

//go:generate ./bin/bin/protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./pkg/proto/url-shortener.proto
