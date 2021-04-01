package minerworker

import (
	"fmt"
	"os"
)

// 开始
func (m *MinerWorker) Start() {

	fmt.Printf("[Start] connect: %s, reward: %s. \n",
		m.config.PoolAddress.String(),
		m.config.Rewards.ToReadable(),
	)

	err := m.powWorker.InitStart() // 初始化
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// 拨号连接
	err = m.startConnect()
	if err != nil {
		fmt.Println(err)
		fmt.Println("[Miner Worker] Reconnection will be initiated in two minutes...")
	}

	go m.loop() // loop

	if m.powWorker == nil {
		fmt.Println("[Miner Worker] ERROR: must call SetPowWorker() first!")
		os.Exit(0)
	}

	// 开始挖矿（投喂）
	go m.powWorker.Excavate(m.miningStuffFeedingCh, m.miningResultCh)

}
