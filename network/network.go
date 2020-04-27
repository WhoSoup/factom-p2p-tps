package network

const NetworkID = 0xf00b47

type Network interface {
	Init(name, port, seed string, bcast int) (func(), error)
	Name() string
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

func (m Metrics) BytesDownF() string {
	return prettyBytes(m.BytesDown) + "/s"
}
func (m Metrics) BytesUpF() string {
	return prettyBytes(m.BytesUp) + "/s"
}
