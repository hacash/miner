package minerserver

import (
	"fmt"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/miner/message"
)

// find block nonce or change coinbase message
// 下发
func (m *MinerServer) Excavate(input interfaces.Block, resCh chan interfaces.Block) {
	// 开始挖掘新的区块
	m.changelock.Lock()
	defer m.changelock.Unlock()

	// 解析挖矿消息
	var err error = nil
	m.penddingBlockMsg, err = message.CreatePendingMiningBlockStuffByBlock(input)
	if err != nil {
		fmt.Println("MinerServer Excavate Error:", err)
		return
	}

	// 挖掘成功上报通道
	m.successMintCh = resCh

	// fmt.Printf("send pending mining block stuff to %d worker of connected with server\n", len(m.allconns))

	// 给所有连接发送挖掘信息
	for _, v := range m.allconns {
		stuffbts := m.penddingBlockMsg.Serialize()
		message.MsgSendToTcpConn(v.conn, message.MinerWorkMsgTypeMiningBlock, stuffbts)
	}

	// ok
}

func (m *MinerServer) StopMining() {

}
