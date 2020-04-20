package network

const NetworkID = 0xf00b47

type Network interface {
	Init(name, port, seed string) error
	Peers() []string
	DeliverMessage(string, []byte)
	ReadMessage() (string, []byte)
	Start()
	FullBroadcastFlag() string
	BroadcastFlag() string
	RandomFlag() string
}
