package minerworker

import (
	"github.com/hacash/core/interfaces"
	"github.com/hacash/miner/message"
	"net"
)

type MinerWorker struct {
	config *MinerWorkerConfig

	conn *net.TCPConn // 连接

	pendingMiningBlockStuff *message.MsgPendingMiningBlockStuff

	miningStuffFeedingCh chan interfaces.PowWorkerMiningStuffItem
	miningResultCh       chan interfaces.PowWorkerMiningStuffItem

	powWorker interfaces.PowWorker // 挖掘器
}

func NewMinerWorker(cnf *MinerWorkerConfig) *MinerWorker {

	worker := &MinerWorker{
		config:               cnf,
		miningStuffFeedingCh: make(chan interfaces.PowWorkerMiningStuffItem, 1),
		miningResultCh:       make(chan interfaces.PowWorkerMiningStuffItem, 1),
	}

	return worker
}

///////////////

// 挖矿执行器
func (m *MinerWorker) SetPowWorker(worker interfaces.PowWorker) {
	m.powWorker = worker
}
