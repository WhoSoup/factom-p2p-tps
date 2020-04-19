package network

import (
	"fmt"
	"sync"
	"time"

	"github.com/FactomProject/factomd/p2p"
	"github.com/WhoSoup/factom-p2p-max/app"
	"github.com/rs/zerolog/log"
)

type V9 struct {
	metrics    chan interface{}
	controller *p2p.Controller

	peerMtx   sync.RWMutex
	connected []string
}

var _ Network = (*V9)(nil)

func NewV9() Network {
	v9 := new(V9)
	return v9
}

func (v9 *V9) consumeMetrics() {
	for m := range v9.metrics {
		if mm, ok := m.(map[string]p2p.ConnectionMetrics); ok {
			v9.peerMtx.Lock()
			v9.connected = make([]string, 0, len(mm))
			for p := range mm {
				v9.connected = append(v9.connected, p)
			}
			v9.peerMtx.Unlock()
		}
	}
}

func (v9 *V9) Start() {
	v9.controller.StartNetwork()
	go v9.consumeMetrics()
}

func (v9 *V9) Init(name, port, seed string) error {
	v9.metrics = make(chan interface{}, p2p.StandardChannelSize)
	p2p.NetworkDeadline = time.Minute * 5
	p2p.CurrentNetwork = app.NetworkID
	p2p.NetworkListenPort = port

	ci := p2p.ControllerInit{
		NodeName:                 name,
		Port:                     port,
		PeersFile:                "",
		Network:                  app.NetworkID,
		Exclusive:                false,
		ExclusiveIn:              false,
		SeedURL:                  seed,
		ConfigPeers:              "",
		CmdLinePeers:             "",
		ConnectionMetricsChannel: v9.metrics,
	}
	v9.controller = new(p2p.Controller).Init(ci)
	return nil
}

func (v9 *V9) Peers() []string { return v9.connected }
func (v9 *V9) DeliverMessage(target string, payload []byte) {
	parc := p2p.NewParcel(app.NetworkID, payload)
	if target == "" {
		target = p2p.BroadcastFlag
	}
	parc.Header.TargetPeer = target
	parc.Header.Type = p2p.TypeMessage

	p2p.BlockFreeChannelSend(v9.controller.ToNetwork, parc)
}
func (v9 *V9) ReadMessage() (string, byte) {
	raw := <-v9.controller.FromNetwork
	if parc, ok := raw.(p2p.Parcel); ok {
		if len(parc.Payload) > 0 {
			return parc.Header.TargetPeer, parc.Payload[0]
		}
		log.Error().Str("peer", parc.Header.TargetPeer).Msg("received empty payload message")
		return parc.Header.TargetPeer, 0
	}
	log.Error().Str("type", fmt.Sprintf("%t", raw)).Msg("received non-parcel message")
	return "", 0
}
