package host

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"tcs/internal/model"
	"tcs/internal/util"
	ws "tcs/internal/websocket"
	"time"

	"github.com/gorilla/websocket"
)

const APPLICATION_NAME = "techcyte-context-sync"
const DEFAULT_TIMEOUT = 30 // In seconds.

type Manager struct {
	Address            string                  // The address we are listening on.
	Clients            map[string]model.Client // A map of client ids to clients.
	SubscribedClientID string                  // The client id for the currently subscribed client.
	disconnect         chan model.Client       // Used to track when clients disconnect.
	Upgrader           websocket.Upgrader      // Used for the websocket connection.
	Context            []model.ContextItem     // The current context.
	VoteContext        []model.ContextItem     // The context in the context change request.

	// For the TUI
	CurrentCase   string   // The case number that is displayed to the user. This is the case number in the current context.
	Voting        bool     // "Voting" in this context means the client has send a context change request and the host has to accept or reject it.
	VoteCase      string   // The case number in the context change request to be accepted or rejected.
	AutoAccept    bool     // If true any context change request will be automatically accepted.
	MessagesToAdd []string // Used for printing to the console in the TUI.
}

func NewManager(address, startingCase string) *Manager {
	// See https://pkg.go.dev/github.com/gorilla/websocket
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return &Manager{
		Address:    address,
		Upgrader:   upgrader,
		Clients:    make(map[string]model.Client),
		disconnect: make(chan model.Client),
		Context: []model.ContextItem{
			{Key: "patient", Value: "p-123456"},
			{Key: "order", Value: "o-654321"},
			{Key: "case", Value: startingCase},
		},
		CurrentCase: startingCase,
	}
}

func Serve(manager *Manager, w http.ResponseWriter, r *http.Request) {
	conn, err := manager.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		manager.PrintErr(err, "error reading message")
		return
	}

	client, err := ws.NewWebsocketClient(manager, conn, msg)
	if err != nil {
		manager.PrintErr(err, "error failed to create new client")
		return
	}
	manager.AddClient(client)

	go client.Read()
	go client.Write()

	manager.ReceiveMessage(client, msg)
}

// Print functions to print to the console in the TUI.
func (m *Manager) Println(msg string) {
	m.MessagesToAdd = append(m.MessagesToAdd, msg)
}

func (m *Manager) Printf(msgFmt string, args ...any) {
	msg := fmt.Sprintf(msgFmt, args...)
	m.MessagesToAdd = append(m.MessagesToAdd, msg)
}

func (m *Manager) PrintErr(err error, msgFmt string, args ...any) {
	msgFmt = strings.Replace(msgFmt, "error ", "", 1)

	msg := fmt.Sprintf(msgFmt, args...)
	msg = fmt.Sprintf("\033[91mError\033[0m %v: %v", msg, err.Error())
	m.MessagesToAdd = append(m.MessagesToAdd, msg)
}

func (m *Manager) PrintErrString(msgFmt string, args ...any) {
	msg := fmt.Sprintf(msgFmt, args...)
	msg = fmt.Sprintf("\033[91mError\033[0m: %v", msg)
	m.MessagesToAdd = append(m.MessagesToAdd, msg)
}

func (m *Manager) AddClient(client model.Client) {
	m.Printf("Application \033[94m'%v'\033[0m connected", client.Application())
	m.Clients[client.ID()] = client
}

func (m *Manager) ClientCount() int {
	return len(m.Clients)
}

func (m *Manager) CaseNumberFromContext(context []model.ContextItem) string {
	if len(context) == 0 {
		return ""
	}

	for _, ctxItem := range context {
		if ctxItem.Key == model.CaseNumber {
			return ctxItem.Value
		}
	}

	return ""
}

func (m *Manager) SetCurrentCaseFromContext() {
	m.CurrentCase = m.CaseNumberFromContext(m.Context)
}

func (m *Manager) HandleMessage(client model.Client, message model.Message) {
	switch message.Kind {
	case model.SubscriptionRequest:
		timeout := (time.Second * DEFAULT_TIMEOUT).Seconds()
		if m.SubscribedClientID != "" {
			message := util.NewSubRejectMessage(APPLICATION_NAME, &timeout, "Already have a subscribed client.", model.ConflictWithRetry)
			m.SendMessage(client, message)

			return
		}

		m.SubscribedClientID = client.ID()
		message := util.NewSubAcceptMessage(APPLICATION_NAME, &timeout, m.CurrentCase)
		m.SendMessage(client, message)
	case model.ContextChangeRequest:
		if len(message.Context) == 0 {
			m.Printf("Empty context on '%v' event.", model.ContextChangeRequest)
			return
		}

		m.VoteContext = message.Context
		m.VoteCase = m.CaseNumberFromContext(message.Context)
		m.Voting = true

		if m.AutoAccept {
			m.Accept()
		}
	case model.ContextChangeAccept:
		m.Context = []model.ContextItem{}
		m.Context = append(m.Context, m.VoteContext...)
		m.CurrentCase = m.VoteCase

		m.VoteContext = []model.ContextItem{}
		m.VoteCase = ""
	case model.ContextChangeReject:
		m.VoteContext = []model.ContextItem{}
		m.VoteCase = ""
	case model.OutOfSyncError:
		m.PrintErrString("Out of sync with client!")
	default:
		m.Printf("Unknown message kind '%v'", message.Kind)
	}
}

func (m *Manager) ReceiveMessage(client model.Client, msg []byte) {
	var message model.Message
	err := json.Unmarshal(msg, &message)
	if err != nil {
		m.PrintErr(err, "error unmarshalling received message: %v", string(msg))
		return
	}

	messageStr, err := util.PrettyPrintMessage(message)
	if err != nil {
		m.PrintErr(err, "error failed to print message on receive")
	} else {
		m.Printf("\033[93mReceived\033[0m message: '%v' from '%v' with payload\n%v", message.Kind, client.Application(), messageStr)
	}

	m.HandleMessage(client, message)
}

func (m *Manager) SendMessage(client model.Client, message model.Message) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		m.PrintErr(err, "error could not marshal message to send")
		return
	}

	messageStr, err := util.PrettyPrintMessage(message)
	if err != nil {
		m.PrintErr(err, "error failed to print message on send")
	} else {
		m.Printf("\033[92mSending\033[0m message: '%v' to '%v' with payload\n%v", message.Kind, client.Application(), messageStr)
	}

	client.SendMessage(messageBytes)
}

func (m *Manager) Accept() {
	m.CurrentCase = m.VoteCase
	m.Context = []model.ContextItem{{Key: model.CaseNumber, Value: m.CurrentCase}}

	m.Voting = false
	m.VoteContext = []model.ContextItem{}
	m.VoteCase = ""

	client := m.Clients[m.SubscribedClientID]
	message := util.NewCtxAcceptMessage(m.Context)
	m.SendMessage(client, message)
}

func (m *Manager) Reject() {
	client := m.Clients[m.SubscribedClientID]
	message := util.NewCtxRejectMessage(m.Context, m.VoteContext, "User rejected context change.", model.BadRequest) // Or other reason.
	m.SendMessage(client, message)

	m.Voting = false
	m.VoteContext = []model.ContextItem{}
	m.VoteCase = ""
}

func (m *Manager) ContextChangeRequest(caseNumber string) {
	if m.SubscribedClientID == "" {
		return
	}

	message := util.NewCtxChangeMessage(caseNumber)
	m.VoteContext = message.Context
	m.VoteCase = caseNumber

	client := m.Clients[m.SubscribedClientID]
	m.SendMessage(client, message)
}

func (m *Manager) ListenForDisconnect() {
	for client := range m.disconnect {
		m.Printf("Application \033[94m'%v'\033[0m disconnected", client.Application())
		delete(m.Clients, client.ID())
		if m.SubscribedClientID == client.ID() {
			m.SubscribedClientID = ""

			// If there are other clients connected pick one to become the new subscribed client.
			// In this example the client we pick is random but it could be done on a FIFO basis.
			// Or a client could be picked for whatever reason.
			for _, nextClient := range m.Clients {
				timeout := (time.Second * DEFAULT_TIMEOUT).Seconds()
				m.SubscribedClientID = nextClient.ID()
				message := util.NewSubAcceptMessage(APPLICATION_NAME, &timeout, m.CurrentCase)
				m.SendMessage(nextClient, message)
				break
			}
		}

		client.Close()
	}
}

func (m *Manager) Disconnect() chan model.Client {
	return m.disconnect
}
