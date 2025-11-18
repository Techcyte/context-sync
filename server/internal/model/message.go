package model

type MessageKind string

const (
	SyncRequest          MessageKind = "sync-request"
	SyncAccept           MessageKind = "sync-accept"
	SyncReject           MessageKind = "sync-reject"
	ContextChangeRequest MessageKind = "ctx-change-request"
	ContextChangeAccept  MessageKind = "ctx-change-accept"
	ContextChangeReject  MessageKind = "ctx-change-reject"
	OutOfSyncError       MessageKind = "sync-error"
)

type ConnectionInfo struct {
	Version              float64  `json:"version"`
	Application          string   `json:"application"`
	Timeout              *float64 `json:"timeout,omitempty"`
	ReplaceExitingClient *bool    `json:"replace_exiting_client,omitempty"`
}

type MessageRejection struct {
	Reason string     `json:"reason"`
	Status StatusCode `json:"status"`
}

type MessageError struct {
	Message string     `json:"message"`
	Status  StatusCode `json:"status"`
}

type Message struct {
	Kind           MessageKind       `json:"kind"`
	Info           *ConnectionInfo   `json:"info,omitempty"`
	Context        []ContextItem     `json:"context,omitempty"`
	CurrentContext []ContextItem     `json:"current_context,omitempty"`
	Rejection      *MessageRejection `json:"rejection,omitempty"`
	Error          *MessageError     `json:"error,omitempty"`
}
