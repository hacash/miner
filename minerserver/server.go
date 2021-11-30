package minerserver

import (
	"github.com/hacash/core/interfacev2"
	"github.com/hacash/miner/message"
	"sync"
)

type MinerServer struct {
	config *MinerServerConfig

	allconns map[uint64]*MinerServerClinet // 全部 TCP 连接

	penddingBlockMsg *message.MsgPendingMiningBlockStuff // 当前正在挖掘的区块消息
	successMintCh    chan interfacev2.Block              // 当前正确挖掘区块的返回

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
func (m *MinerServer) Start() error {
	return m.startListen()
}
