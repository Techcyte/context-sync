package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"tcs/internal/certs"
	"tcs/internal/server"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	port := flag.String("port", "4002", "What port to use")
	startingCase := flag.String("case", "N123456", "Starting case number")
	autoAccept := flag.Bool("auto-accept", false, "If enabled the manager will auto accept context change requests")
	flag.Parse()

	// Generate and trust a self-signed cert if we don't have
	// one yet. This runs before the TUI starts.
	if !certs.CertificatesExist(".") {
		fmt.Println("No TLS certificate found. Generating a self-signed certificate...")
	}
	generated, warning, err := certs.Ensure(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate a TLS certificate: %v\n", err)
		os.Exit(1)
	}
	if generated {
		fmt.Printf("Generated %s and %s.\n", certs.ServerCertFile, certs.ServerKeyFile)
		if warning != nil {
			fmt.Printf("Could not install the CA into the trust store automatically: %v", warning)
			fmt.Printf("The server will still start, but you must trust %s manually. See the README.\n", certs.CACertFile)
			fmt.Println("Press Enter to start.")
			fmt.Scanln()
		} else {
			switch runtime.GOOS {
			case "windows":
				fmt.Printf("Installed %s into the trust store.\n", certs.CACertFile);
			case "darwin":
				fmt.Printf("Installed %s into the trust store. On macOS, open Keychain Access and set it to \"Always Trust\".\n", certs.CACertFile)
			}
			fmt.Println("Press Enter to start.")
			fmt.Scanln()
		}
	}

	address := fmt.Sprintf(":%v", *port)
	manager := server.NewManager(address, *startingCase)
	if autoAccept != nil {
		manager.AutoAccept = *autoAccept
	}

	go manager.ListenForDisconnect()
	http.HandleFunc("/cm", func(w http.ResponseWriter, r *http.Request) {
		server.Serve(manager, w, r)
	})

	// Remove tea.WithAltScreen() to NewProgram() if you want to retain the text on screen after the program exits.
	application := server.NewApp(manager)
	_, err = tea.NewProgram(application, tea.WithAltScreen(), tea.WithMouseAllMotion()).Run()
	if err != nil {
		panic(err)
	}
}
