package minerrelayservice

import (
	"github.com/hacash/miner/message"
	"net"
	"sync"
)

type RelayService struct {
	config *MinerRelayServiceConfig

	service_tcp *net.TCPConn

	changelock sync.Mutex

	allconns map[uint64]*ConnClient // 全部 TCP 连接

	penddingBlockStuff *message.MsgPendingMiningBlockStuff // 当前正在挖掘的区块消息
	//successMintCh    chan interfaces.Block               // 当前正确挖掘区块的返回

}

func NewRelayService(cnf *MinerRelayServiceConfig) *RelayService {
	return &RelayService{
		config:             cnf,
		service_tcp:        nil,
		allconns:           make(map[uint64]*ConnClient),
		penddingBlockStuff: nil,
	}
}

func (r *RelayService) Start() {

	go r.startListen()

	go r.connectToService()

	go r.loop()

}
