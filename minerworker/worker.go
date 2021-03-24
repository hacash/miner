package minerworker

import (
	"fmt"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/miner/message"
	"net"
	"os"
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

// 开始
func (m *MinerWorker) Start() {

	fmt.Printf("[Start] connect: %s, rewards: %s, supervene: %d. \n",
		m.config.PoolAddress.String(),
		m.config.Rewards.ToReadable(),
		m.config.Supervene,
	)

	// 拨号连接
	err := m.startConnect()
	if err != nil {
		fmt.Println(err)
		fmt.Println("[Miner Worker] Reconnection will be initiated in two minutes...")
	}

	go m.loop()

	if m.powWorker == nil {
		fmt.Println("[Miner Worker] ERROR: must call SetPowWorker() first!")
		os.Exit(0)
	}

	err = m.powWorker.InitStart() // 初始化
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// 开始挖矿（投喂）
	go m.powWorker.Excavate(m.miningStuffFeedingCh, m.miningResultCh)

}

// 挖矿执行器
func (m *MinerWorker) SetPowWorker(worker interfaces.PowWorker) {
	m.powWorker = worker
}
