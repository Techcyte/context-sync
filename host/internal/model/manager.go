package model

type Manager interface {
	Println(msg string)
	Printf(msgFmt string, args ...any)
	PrintErr(err error, msgFmt string, args ...any)
	ReceiveMessage(client Client, msg []byte)
	SendMessage(client Client, message Message)
	Disconnect() chan Client
}
