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

	stats *Stats
}

type Stats struct {
	mtx             sync.RWMutex
	NonDupeMessages []uint64
	Messages        []uint64
	Sent            []uint64
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

	a.gen = new(Generator)

	rand.Seed(time.Now().UnixNano())
	a.replay = NewReplay(time.Minute, 10)
	return a
}

func (a *App) SetNetwork(n network.Network) {
	a.n = n
}

func (a *App) Stats() *Stats {
	return a.stats
}

func (a *App) worker() {
	for {
		peer, msg := a.n.ReadMessage()
		if len(msg) == 0 {
			log.Warn().Str("peer", peer).Msg("received invalid message")
			continue
		}

		hash := sha256.Sum256(msg)
		a.stats.mtx.Lock()
		a.stats.Messages[msg[0]]++
		a.stats.mtx.RUnlock()

		if !a.replay.Dupe(fmt.Sprintf("%x", hash)) {

			sent := byte(0)
			switch msg[0] {
			case ACK, EOM, Heartbeat, CommitChain, CommitEntry, RevealEntry, DBSig, Transaction: // rebroadcast
				a.n.DeliverMessage(a.n.BroadcastFlag(), msg)
				sent = msg[0]
			case MissingMsg: // reply
				a.n.DeliverMessage(peer, a.gen.CreateMessage(MissingReply))
				sent = MissingReply
			case DBStateRequest:
				a.n.DeliverMessage(peer, a.gen.CreateMessage(DBStateReply))
				sent = DBStateReply
			case MissingReply, DBStateReply:
				// ignore
			default:
				log.Warn().Str("peer", peer).Int("len", len(msg)).Msg("received invalid message with payload")
			}
			a.stats.mtx.Lock()
			a.stats.NonDupeMessages[msg[0]]++
			if sent > 0 {
				a.stats.Sent[sent]++
			}
			a.stats.mtx.Unlock()
		}

	}
}

func (a *App) Launch() {

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
	for i := 0; i < a.feds; i++ {
		a.n.DeliverMessage(a.n.RandomFlag(), a.gen.CreateMessage(EOM))
	}
	for i := 0; i < a.audits; i++ {
		a.n.DeliverMessage(a.n.RandomFlag(), a.gen.CreateMessage(Heartbeat))
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
		log.Info().Int("eps", eps).Int("feds", feds).Int("audits", audits).Msg("enabling load generator")
	} else {
		log.Info().Msg("load generating disabled")
	}
}
