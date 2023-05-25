package minerserver

import (
	"github.com/hacash/core/interfaces"
	interfaces2 "github.com/hacash/miner/interfaces"
	"sync"
)

type MinerServer struct {
	config *MinerServerConfig

	allconns map[uint64]*MinerServerClient // All TCP connections

	penddingBlockMsg *interfaces2.PoWStuffOverallData // Currently mining block messages
	successMintCh    chan interfaces.Block            // Return of currently correct mining block

	changelock sync.Mutex
}

// new
func NewMinerServer(cnf *MinerServerConfig) *MinerServer {

	serv := &MinerServer{
		config:   cnf,
		allconns: make(map[uint64]*MinerServerClient),
	}

	return serv
}

// start
func (m *MinerServer) Start() error {
	return m.startListen()
}
