package network

import (
	"fmt"
	"math/rand"
	"time"

	p2p "github.com/WhoSoup/factom-p2p"
	"github.com/rs/zerolog/log"
	"github.com/sirupsen/logrus"
)

type V10 struct {
	config p2p.Configuration
	n      *p2p.Network

	metrics   Metrics
	connected []string
}

var _ Network = (*V10)(nil)

func NewV10(version int) Network {
	v10 := new(V10)
	v10.config = p2p.DefaultP2PConfiguration()
	v10.config.ChannelCapacity = 10000
	v10.config.ProtocolVersion = uint16(version)
	if version == 11 {
		logrus.SetLevel(logrus.DebugLevel)
	}
	return v10
}

func (v10 *V10) Metrics() Metrics {
	return v10.metrics
}
func (v10 *V10) processMetrics() {
	ticker := time.NewTicker(time.Second)
	old := make(map[string]p2p.PeerMetrics)

	for range ticker.C {
		nm := Metrics{}
		newm := v10.n.GetPeerMetrics()
		connected := make([]string, 0, len(newm))
		for hash, m := range newm {
			connected = append(connected, hash)
			// check if new peer
			if oldmetrics, ok := old[hash]; ok {
				nm.BytesDown += m.BytesReceived - oldmetrics.BytesReceived
				nm.BytesUp += m.BytesSent - oldmetrics.BytesSent
				nm.MessagesDown += m.MessagesReceived - oldmetrics.MessagesReceived
				nm.MessagesUp += m.MessagesSent - oldmetrics.MessagesSent
			} else {
				nm.BytesDown += m.BytesReceived
				nm.BytesUp += m.BytesSent
				nm.MessagesDown += m.MessagesReceived
				nm.MessagesUp += m.MessagesSent
			}
		}
		old = newm
		v10.metrics = nm
		v10.connected = connected
	}
}
func (v10 *V10) Name() string {
	return fmt.Sprintf("%s-%d", v10.config.NodeName, v10.config.NodeID)
}
func (v10 *V10) Start() {
	go v10.processMetrics()
	log.Fatal().Err(v10.n.Run())
}
func (v10 *V10) Init(name, port, seed string, bcast int) (func(), error) {
	v10.config.NodeName = name
	v10.config.SeedURL = seed
	v10.config.ListenPort = port
	v10.config.Fanout = uint(bcast)
	v10.config.NodeID = rand.Uint32()
	nn, err := p2p.NewNetwork(v10.config)
	if err != nil {
		return nil, err
	}
	v10.n = nn
	return func() {}, nil
}
func (v10 *V10) Peers() []string {
	return v10.connected
}
func (v10 *V10) DeliverMessage(target string, payload []byte) {
	parc := p2p.NewParcel(target, payload)
	v10.n.Send(parc)
}

func (v10 *V10) ReadMessage() (string, []byte) {
	p := <-v10.n.Reader()
	if len(p.Payload) == 0 {
		log.Error().Str("peer", p.Address).Msg("received empty payload message")
		return p.Address, nil
	}
	return p.Address, p.Payload
}

func (v10 *V10) FullBroadcastFlag() string { return p2p.FullBroadcast }
func (v10 *V10) BroadcastFlag() string     { return p2p.Broadcast }
func (v10 *V10) RandomFlag() string        { return p2p.RandomPeer }
