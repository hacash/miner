package minerserver

import (
	"github.com/hacash/core/interfaces"
	"github.com/hacash/miner/message"
	"sync"
)

type MinerServer struct {
	config *MinerServerConfig

	allconns map[uint64]*MinerServerClinet // All TCP connections

	penddingBlockMsg *message.MsgPendingMiningBlockStuff // Currently mining block messages
	successMintCh    chan interfaces.Block               // Return of currently correct mining block

	changelock sync.Mutex
}

// new
func NewMinerServer(cnf *MinerServerConfig) *MinerServer {

	serv := &MinerServer{
		config:   cnf,
		allconns: make(map[uint64]*MinerServerClinet),
	}

	return serv
}

// start
func (m *MinerServer) Start() error {
	return m.startListen()
}
