package minerserver

import (
	"fmt"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/miner/message"
	"os"
	"sync"
)

type MinerServer struct {
	config *MinerServerConfig

	allconns map[uint64]*MinerServerClinet // 全部 TCP 连接

	penddingBlockMsg *message.MsgPendingMiningBlockStuff // 当前正在挖掘的区块消息
	successMintCh    chan interfaces.Block               // 当前正确挖掘区块的返回

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

// 开始
func (m *MinerServer) Start() {
	go func() {
		err := m.startListen()
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}()
}
