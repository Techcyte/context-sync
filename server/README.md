# Server

The demo server is a `go` project so you will need to have go installed in order to use it. Go can be installed from [here](https://go.dev/doc/install).

### macOS / Linux

To build the demo run `make build` or `go mod tidy && go build -o techcyte_context_sync_host cmd/tcs/main.go`.

To build and run the demo run `make br` or `go mod tidy && go build -o techcyte_context_sync_host cmd/tcs/main.go && ./techcyte_context_sync_host`.

### Windows

Use the batch scripts in this folder (double-click them or run them from a command prompt or terminal):

* `build.bat` — builds `techcyte_context_sync_host.exe`.
* `run.bat` — builds and then runs it.

Or run the equivalent commands directly:

```
go mod tidy
go build -o techcyte_context_sync_host.exe cmd\tcs\main.go
techcyte_context_sync_host.exe
```

## Keyboard hotkeys

The demo server is a TUI app. Press `n` to input a new case number then press `enter` to send a context change request. Press `c` to clear the console. Press `q` to quit.
For incoming context change requests press `a` to accept and `r` to reject.

## Secure WebSockets and self-signed certificates

The LIS Protocol runs over **secure WebSockets (`wss://`)**, which is just a WebSocket over TLS, the same encryption a browser uses for `https://`. Because Fusion (the client) is a web application, browsers will refuse to open a `ws://` (unencrypted) connection from a secure page, so the server **must** serve `wss://`.

Serving TLS requires a **certificate** and a **private key**. In production you would use a certificate issued by a certificate authority (CA) that browsers already trust. For local development that isn't practical, there's no public hostname to issue a certificate for, so we use a **self-signed certificate** instead.

### What the demo server does automatically

The first time you run the server it generates its own certificates using nothing but the Go standard library, so **no extra software is required on Windows or macOS**. On startup, if the certificate files are missing, the server:

1. Generates a small local **certificate authority** — `ca.crt` (the certificate) and `ca.key` (its private key).
2. Generates a **server certificate** for `localhost` — `server.crt` and `server.key` — signed by that CA. It includes `localhost` and `127.0.0.1` as Subject Alternative Names so the browser accepts either.
3. Attempts to **install the CA into your operating system's trust store** so the browser will trust the server certificate.

These files are written to the directory you run the server from and are git-ignored. Delete them and restart the server to regenerate a fresh set.

> **Why a CA plus a server certificate, instead of one self-signed certificate?**
> You trust the CA (`ca.crt`) **once**. After that, any server certificate the CA signs is trusted automatically, so you can regenerate `server.crt` as often as you like without touching the trust store again. Trusting the CA, not an individual server certificate, is also how browsers are designed to work.

### Trusting the CA

For the browser to accept the connection, `ca.crt` has to be trusted. The server tries to do this for you, but the exact behavior differs by platform:

**Windows** — The server runs `certutil -user -addstore Root ca.crt`, which adds the CA to the current user's *Trusted Root Certification Authorities* store. The first time, Windows shows a security dialog asking you to confirm, click **Yes**. No administrator rights are needed. To do it manually:

```
certutil -user -addstore Root ca.crt
```

**macOS** — The server adds `ca.crt` to your login keychain and opens **Keychain Access**. macOS does not allow a program to flip the *trust* setting without your consent, so you must do the last step yourself: in Keychain Access, find **Techcyte Local CA**, double-click it, expand **Trust**, and set **When using this certificate** to **Always Trust**. Alternatively, on macOS you can use [`mkcert`](https://github.com/FiloSottile/mkcert) (`brew install mkcert && mkcert -install`) to manage a trusted local CA, though it is not required.

After the CA is trusted, restart the browser tab and connect to `wss://localhost:4002/cm`. If you skip this step the browser will silently fail to open the WebSocket connection (there is no "click to proceed" prompt for WebSockets like there is for a web page).
