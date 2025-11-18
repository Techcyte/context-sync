package main

import (
	"flag"
	"fmt"
	"net/http"
	"tcs/internal/server"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	port := flag.String("port", "4002", "What port to use")
	startingCase := flag.String("case", "N123456", "Starting case number")
	autoAccept := flag.Bool("auto-accept", false, "If enabled the manager will auto accept context change requests")
	flag.Parse()

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
	_, err := tea.NewProgram(application, tea.WithAltScreen(), tea.WithMouseAllMotion()).Run()
	if err != nil {
		panic(err)
	}
}
