package network

type Network interface {
	Init(name, port, seed string) error
	Peers() []string
	DeliverMessage(string, []byte)
	ReadMessage() (string, byte)
	Start()
}
