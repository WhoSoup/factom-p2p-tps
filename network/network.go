package network

const NetworkID = 0xf00b47

type Network interface {
	Init(name, port, seed string) error
	Peers() []string
	Metrics() Metrics
	DeliverMessage(string, []byte)
	ReadMessage() (string, []byte)
	Start()
	FullBroadcastFlag() string
	BroadcastFlag() string
	RandomFlag() string
}

type Metrics struct {
	BytesDown    uint64
	BytesUp      uint64
	MessagesDown uint64
	MessagesUp   uint64
}
