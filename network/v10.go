package network

import (
	p2p "github.com/WhoSoup/factom-p2p"
	"github.com/rs/zerolog/log"
)

type V10 struct {
	config p2p.Configuration
	n      *p2p.Network
}

var _ Network = (*V10)(nil)

func NewV10(version int) Network {
	v10 := new(V10)
	v10.config = p2p.DefaultP2PConfiguration()
	v10.config.ProtocolVersion = uint16(version)
	return v10
}

func (v10 *V10) Start() {
	log.Fatal().Err(v10.n.Run())
}
func (v10 *V10) Init(name, port, seed string) error {
	v10.config.NodeName = name
	v10.config.SeedURL = seed
	v10.config.ListenPort = port
	nn, err := p2p.NewNetwork(v10.config)
	if err != nil {
		return err
	}
	v10.n = nn
	return nil
}
func (v10 *V10) Peers() []string {
	var peers []string
	for _, m := range v10.n.GetPeerMetrics() {
		peers = append(peers, m.Hash)
	}
	return peers
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
