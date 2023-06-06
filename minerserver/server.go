package minerserver

import (
	"github.com/hacash/core/interfaces"
	interfaces2 "github.com/hacash/miner/interfaces"
	"sync"
)

type MinerServer struct {
	Conf *MinerServerConfig

	allconns map[uint64]*MinerServerClient // All TCP connections

	penddingBlockMsg *interfaces2.PoWStuffOverallData // Currently mining block messages
	successMintCh    chan interfaces.Block            // Return of currently correct mining block

	changelock sync.Mutex
}

// new
func NewMinerServer(cnf *MinerServerConfig) *MinerServer {

	serv := &MinerServer{
		Conf:     cnf,
		allconns: make(map[uint64]*MinerServerClient),
	}

	return serv
}

func (m *MinerServer) Config() interfaces2.PoWConfig {
	return m.Conf
}

// start
func (m *MinerServer) Start() error {
	return m.startListen()
}
