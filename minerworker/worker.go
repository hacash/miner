package minerworker

import (
	"github.com/hacash/core/interfacev2"
	"github.com/hacash/miner/message"
	"net"
)

type MinerWorker struct {
	config *MinerWorkerConfig

	conn *net.TCPConn // 连接

	pendingMiningBlockStuff *message.MsgPendingMiningBlockStuff

	miningStuffFeedingCh chan interfacev2.PowWorkerMiningStuffItem
	miningResultCh       chan interfacev2.PowWorkerMiningStuffItem

	powWorker interfacev2.PowWorker // 挖掘器
}

func NewMinerWorker(cnf *MinerWorkerConfig) *MinerWorker {

	worker := &MinerWorker{
		config:               cnf,
		miningStuffFeedingCh: make(chan interfacev2.PowWorkerMiningStuffItem, 1),
		miningResultCh:       make(chan interfacev2.PowWorkerMiningStuffItem, 1),
	}

	return worker
}

///////////////

// 挖矿执行器
func (m *MinerWorker) SetPowWorker(worker interfacev2.PowWorker) {
	m.powWorker = worker
}
