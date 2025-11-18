package model

type Client interface {
	SendMessage([]byte)
	Close()

	ID() string
	Application() string

	SetTransaction(string)
	Transaction() string
}
