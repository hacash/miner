package minerworker

import (
	"fmt"
	"os"
)

// start
func (m *MinerWorker) Start() {

	fmt.Printf("[Start] connect: %s, reward: %s. \n",
		m.config.PoolAddress.String(),
		m.config.Rewards.ToReadable(),
	)

	//err := m.powMaster.Init() // 初始化
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(0)
	//}

	// Dial up connection
	err := m.startConnect()
	if err != nil {
		fmt.Println(err)
		fmt.Println("[Miner Worker] Reconnection will be initiated in two minutes...")
	}

	go m.loop() // loop

	if m.powWorker == nil {
		fmt.Println("[Miner Worker] ERROR: must call SetPowWorker() first!")
		os.Exit(0)
	}

	// Start mining (feeding)
	//go m.Excavate(m.miningStuffFeedingCh, m.miningResultCh)

}
