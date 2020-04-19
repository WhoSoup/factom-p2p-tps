package app

import "time"

// message ID
const (
	Invalid int = iota
	ACK
	EOM
	Heartbeat
	CommitChain
	CommitEntry
	RevealEntry
	DBSig
	Transaction
	MissingMsg
	MissingReply
	DBStateRequest
	DBStateReply
)

const NetworkID = 0xf00b47

// Average Byte-Size of messages
// calculated from 68 hours of mainnet traffic
var avgSize = map[int]int{
	ACK:            256,
	EOM:            179,
	Heartbeat:      175,
	CommitChain:    201,
	CommitEntry:    137,
	RevealEntry:    538,
	DBSig:          385,
	Transaction:    250,
	MissingMsg:     56,
	MissingReply:   538,
	DBStateRequest: 15,
	DBStateReply:   785,
}

var minuteDuration = time.Minute
var blockDuration = 10 * minuteDuration

var dbstateLikelihood = 0.7621359223300971    // 76.2% likelihood for dbstate request after block duration, 314 / 412
var missingmsgLikelihood = 0.7008092142418409 // 70.1% likelihood of missingmsg for every *NEW* ACK, 170003 / 242581

// makeup of transactions to chains to entries
var entryPercent = map[int]float64{
	CommitChain: 0.0076831142222981,
	Transaction: 0.0012975926242103,
	CommitEntry: 0.9910192931534915,
}
