package network

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/p2p"
	"github.com/rs/zerolog/log"
)

type V9 struct {
	metricsConsumer chan interface{}
	controller      *p2p.Controller

	connected []string
	metrics   Metrics
}

var _ Network = (*V9)(nil)

func NewV9() Network {
	v9 := new(V9)
	return v9
}

func (v9 *V9) consumeMetrics() {
	last := time.Now()
	old := make(map[string]p2p.ConnectionMetrics)
	for m := range v9.metricsConsumer {
		if mm, ok := m.(map[string]p2p.ConnectionMetrics); ok {
			secs := time.Since(last).Seconds()
			last = time.Now()
			var metrics Metrics
			connected := make([]string, 0, len(mm))
			for p, info := range mm {
				connected = append(connected, p)
				if oldinfo, ok := old[p]; ok {
					metrics.BytesDown += uint64(info.BytesReceived - oldinfo.BytesReceived)
					metrics.BytesUp += uint64(info.BytesSent - oldinfo.BytesSent)
					metrics.MessagesDown += uint64(info.MessagesReceived - oldinfo.MessagesReceived)
					metrics.MessagesUp += uint64(info.MessagesSent - oldinfo.MessagesSent)
				} else {
					metrics.BytesDown += uint64(info.BytesReceived)
					metrics.BytesUp += uint64(info.BytesSent)
					metrics.MessagesDown += uint64(info.MessagesReceived)
					metrics.MessagesUp += uint64(info.MessagesSent)
				}
			}

			// normalize to 1s
			metrics.BytesDown = uint64(float64(metrics.BytesDown) / secs)
			metrics.BytesUp = uint64(float64(metrics.BytesUp) / secs)
			metrics.MessagesDown = uint64(float64(metrics.MessagesDown) / secs)
			metrics.MessagesUp = uint64(float64(metrics.MessagesUp) / secs)

			v9.connected = connected
		}
	}
}

func (v9 *V9) Metrics() Metrics {
	return v9.metrics
}

func (v9 *V9) Start() {
	v9.controller.StartNetwork()
	go v9.consumeMetrics()
}

func (v9 *V9) Init(name, port, seed string) error {
	v9.metricsConsumer = make(chan interface{}, p2p.StandardChannelSize)
	p2p.NetworkDeadline = time.Minute * 5
	p2p.CurrentNetwork = NetworkID
	p2p.NetworkListenPort = port

	ci := p2p.ControllerInit{
		NodeName:                 name,
		Port:                     port,
		PeersFile:                "",
		Network:                  NetworkID,
		Exclusive:                false,
		ExclusiveIn:              false,
		SeedURL:                  seed,
		ConfigPeers:              "",
		CmdLinePeers:             "",
		ConnectionMetricsChannel: v9.metricsConsumer,
	}
	v9.controller = new(p2p.Controller).Init(ci)
	return nil
}

func (v9 *V9) Peers() []string { return v9.connected }
func (v9 *V9) DeliverMessage(target string, payload []byte) {
	// we just need msg.GetMsgHash().Fixed(), nothing else
	// ack caches its msghash so it works here
	ack := new(messages.Ack)
	sha := sha256.Sum256(payload)
	ack.MsgHash = primitives.NewHash(sha[:])

	parc := p2p.NewParcelMsg(NetworkID, payload, ack)
	if target == "" {
		target = p2p.BroadcastFlag
	}
	parc.Header.TargetPeer = target
	parc.Header.Type = p2p.TypeMessage

	p2p.BlockFreeChannelSend(v9.controller.ToNetwork, *parc)
}
func (v9 *V9) ReadMessage() (string, []byte) {
	raw := <-v9.controller.FromNetwork
	if parc, ok := raw.(p2p.Parcel); ok {
		if len(parc.Payload) > 0 {
			return parc.Header.TargetPeer, parc.Payload
		}
		log.Error().Str("peer", parc.Header.TargetPeer).Msg("received empty payload message")
		return parc.Header.TargetPeer, nil
	}
	log.Error().Str("type", fmt.Sprintf("%t", raw)).Msg("received non-parcel message")
	return "", nil
}

func (v9 *V9) FullBroadcastFlag() string { return p2p.FullBroadcastFlag }
func (v9 *V9) BroadcastFlag() string     { return p2p.BroadcastFlag }
func (v9 *V9) RandomFlag() string        { return p2p.RandomPeerFlag }
