.PHONY: run
run:
	go run cmd/tcs/main.go

.PHONY: build
build:
	go mod tidy && go build -o techcyte_context_sync_host cmd/tcs/main.go

.PHONY: build_windows
build_windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o techcyte_context_sync_host.exe cmd/tcs/main.go

# build and run
.PHONY: br
br: build
	./techcyte_context_sync_host
