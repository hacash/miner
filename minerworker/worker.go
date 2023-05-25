package minerworker

import (
	interfaces2 "github.com/hacash/miner/interfaces"
	"net"
	"sync"
)

type MinerWorker struct {
	config *MinerWorkerConfig

	conn        *net.TCPConn // connect
	statusMutex sync.Mutex

	//pendingMiningBlockStuff *interfaces2.PoWStuffOverallData

	//miningStuffFeedingCh chan *interfaces2.PoWStuffOverallData
	//miningResultCh       chan *interfaces2.PoWResultData

	powWorker interfaces2.PoWWorker // Digger
}

func NewMinerWorker(cnf *MinerWorkerConfig) *MinerWorker {

	worker := &MinerWorker{
		config: cnf,
		//miningStuffFeedingCh: make(chan *interfaces2.PoWStuffOverallData, 1),
		//miningResultCh:       make(chan *interfaces2.PoWResultData, 1),
	}

	return worker
}

///////////////

// Mining actuator
func (m *MinerWorker) SetPoWWorker(worker interfaces2.PoWWorker) {
	m.powWorker = worker
}
