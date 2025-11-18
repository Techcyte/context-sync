# Server

The demo server is a `go` project so you will need to have go installed in order to use it. Go can be installed from [here](https://go.dev/doc/install).

To build the demo run `make build` or `go mod tidy && go build -o techcyte_context_sync_host cmd/tcs/main.go`.

If you are using Windows run `make build_windows` or `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o techcyte_context_sync_host.exe cmd/tcs/main.go`.

To build and run the demo run `make br` or `go mod tidy && go build -o techcyte_context_sync_host cmd/tcs/main.go && ./techcyte_context_sync_host`.

The demo server is a TUI app. Press `n` to input a new case number then press `enter` to send a context change request. Press `c` to clear the console. Press `q` to quit.
