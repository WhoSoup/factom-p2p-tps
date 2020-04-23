package app

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/WhoSoup/factom-p2p-tps/network"
	"github.com/rs/zerolog/log"
)

type App struct {
	n      network.Network
	gen    *Generator
	replay *Replay

	Height int
	Minute int
	mtx    sync.RWMutex

	generate          bool
	feds, audits, eps int

	loadchange chan int
	loadcancel func()

	stats *Stats
}

type Stats struct {
	mtx             sync.RWMutex
	NonDupeMessages []uint64
	Messages        []uint64
	Sent            []uint64

	TPS      uint64
	TPSCount uint64
	EPS      uint64
	EPSCount uint64

	Metrics network.Metrics
}

func (s *Stats) AddMsg(msg byte, dupe bool) {
	s.mtx.Lock()
	s.Messages[msg]++
	if !dupe {
		s.NonDupeMessages[msg]++
	}
	s.mtx.Unlock()
}

func (s *Stats) AddSent(msg byte, count uint64) {
	s.mtx.Lock()
	s.Sent[msg] += count
	s.mtx.Unlock()
}

func (s *Stats) AddPS(eps, tps uint64) {
	s.mtx.Lock()
	s.TPSCount += tps
	s.EPSCount += eps
	s.mtx.Unlock()
}

func (s *Stats) Waste(b int) float64 {
	if s.Messages[b] == 0 {
		return 0
	}
	return float64(s.NonDupeMessages[b]) / float64(s.Messages[b])
}

func (s *Stats) Name(b int) string {
	return MessageName(b)
}

func NewApp() *App {
	a := new(App)
	a.stats = new(Stats)
	a.stats.Messages = make([]uint64, MESSAGEMAX)
	a.stats.Sent = make([]uint64, MESSAGEMAX)
	a.stats.NonDupeMessages = make([]uint64, MESSAGEMAX)
	a.loadchange = make(chan int)

	a.gen = NewGenerator(entryPercent)

	rand.Seed(time.Now().UnixNano())
	a.replay = NewReplay(time.Minute, 10)
	return a
}

func (a *App) Stats() *Stats {
	a.stats.mtx.Lock()
	defer a.stats.mtx.Unlock()
	if a.n == nil {
		return &Stats{}
	}
	a.stats.Metrics = a.n.Metrics()
	return a.stats
}

func (a *App) generateLoad() {
	for l := range a.loadchange {
		if a.loadcancel != nil {
			a.loadcancel()
			a.loadcancel = nil
		}

		if l <= 0 {
			log.Info().Msg("stopping load gen")
			continue
		}

		stopper := make(chan interface{})
		a.loadcancel = func() {
			close(stopper)
		}

		go func(eps int) {
			t := time.Second / time.Duration(eps)
			log.Info().Int("eps", eps).Dur("interval", t).Msg("starting load gen")
			defer log.Info().Int("eps", eps).Dur("interval", t).Msg("ending load gen")
			ticker := time.NewTicker(t)
			for range ticker.C {
				select {
				case <-stopper:
					return
				default:
				}

				mtype := a.gen.WeightedRandomType()
				a.n.DeliverMessage(a.n.RandomFlag(), a.gen.CreateMessage(mtype))
				a.n.DeliverMessage(a.n.RandomFlag(), a.gen.CreateMessage(ACK))

				if mtype != Transaction {
					a.n.DeliverMessage(a.n.RandomFlag(), a.gen.CreateMessage(RevealEntry))
					a.n.DeliverMessage(a.n.RandomFlag(), a.gen.CreateMessage(ACK))
				}
			}
		}(l)
	}
}

func (a *App) worker() {
	for {
		peer, msg := a.n.ReadMessage()
		if len(msg) == 0 {
			log.Warn().Str("peer", peer).Msg("received invalid message")
			continue
		}

		hash := sha256.Sum256(msg)
		if a.replay.Dupe(fmt.Sprintf("%x", hash)) {
			a.stats.AddMsg(msg[0], true)
		} else {
			sent := byte(0)
			switch msg[0] {
			case ACK, EOM, Heartbeat, CommitChain, CommitEntry, RevealEntry, DBSig, Transaction: // rebroadcast
				a.n.DeliverMessage(a.n.BroadcastFlag(), msg)
				sent = msg[0]
			case MissingMsg: // rebroadcast and reply
				a.n.DeliverMessage(a.n.BroadcastFlag(), msg)
				a.n.DeliverMessage(peer, a.gen.CreateMessage(MissingReply))
				sent = MissingReply
			case DBStateRequest:
				a.n.DeliverMessage(a.n.BroadcastFlag(), msg)
				a.n.DeliverMessage(peer, a.gen.CreateMessage(DBStateReply))
				sent = DBStateReply
			case MissingReply, DBStateReply:
				// ignore
			default:
				log.Warn().Str("peer", peer).Int("len", len(msg)).Msg("received invalid message with payload")
			}
			a.stats.AddMsg(msg[0], false)
			if sent != 0 {
				a.stats.AddSent(sent, 1)
			}

			switch msg[0] {
			case CommitChain, CommitEntry:
				a.stats.AddPS(0, 1)
			case RevealEntry, Transaction:
				a.stats.AddPS(1, 1)
			}

			if a.generate && msg[0] == ACK && rand.Float64() < missingmsgLikelihood {
				a.n.DeliverMessage(a.n.RandomFlag(), a.gen.CreateMessage(MissingMsg))
			}
		}

	}
}

func (a *App) calculateStats() {
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		a.stats.mtx.Lock()
		a.stats.EPS = a.stats.EPSCount
		a.stats.EPSCount = 0
		a.stats.TPS = a.stats.TPSCount
		a.stats.TPSCount = 0
		a.stats.mtx.Unlock()
	}
}

func (a *App) Launch(n network.Network) {
	a.n = n

	go a.generateLoad()
	go a.calculateStats()
	for i := 0; i < workers; i++ {
		go a.worker()
	}

	ticker := time.NewTicker(minuteDuration)
	for range ticker.C {
		a.mtx.Lock()
		a.Minute++
		if a.Minute >= minutesPerBlock {
			a.Height++
			a.Minute = 0
		}
		a.mtx.Unlock()

		if a.generate {
			a.sendEOMs()
		}
	}
}

func (a *App) Settings() (bool, int, int, int) {
	a.mtx.RLock()
	defer a.mtx.RUnlock()
	return a.generate, a.eps, a.feds, a.audits
}

func (a *App) sendEOMs() {
	a.mtx.RLock()
	defer a.mtx.RUnlock()
	// seed these out to random peers first
	typ := EOM
	if a.Minute == 0 {
		typ = DBSig
	}
	for i := 0; i < a.feds; i++ {
		a.n.DeliverMessage(a.n.RandomFlag(), a.gen.CreateMessage(typ))
	}
	for i := 0; i < a.audits; i++ {
		a.n.DeliverMessage(a.n.RandomFlag(), a.gen.CreateMessage(Heartbeat))
	}

	if a.Minute == 0 && rand.Float64() < dbstateLikelihood {
		a.n.DeliverMessage(a.n.RandomFlag(), a.gen.CreateMessage(DBStateRequest))
	}
}

func (a *App) ApplyLoad(generate bool, eps, feds, audits int) {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.generate = generate
	a.eps = eps
	a.feds = feds
	a.audits = audits

	if generate {
		log.Info().Int("eps", eps).Int("feds", feds).Int("audits", audits).Msg("setting load generator")
		a.loadchange <- eps
	} else {
		log.Info().Msg("load generating disabled")
		a.loadchange <- 0
	}
}
