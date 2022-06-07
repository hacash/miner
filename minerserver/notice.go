package minerserver

import (
	"fmt"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/miner/message"
)

// find block nonce or change coinbase message
// Issue
func (m *MinerServer) Excavate(input interfaces.Block, resCh chan interfaces.Block) {
	// Start mining new blocks
	m.changelock.Lock()
	defer m.changelock.Unlock()

	// Parsing mining messages
	var err error = nil
	m.penddingBlockMsg, err = message.CreatePendingMiningBlockStuffByBlock(input.(interfaces.Block))
	if err != nil {
		fmt.Println("MinerServer Excavate Error:", err)
		return
	}

	// Mining successful reporting channel
	m.successMintCh = resCh

	// fmt.Printf("send pending mining block stuff to %d worker of connected with server\n", len(m.allconns))

	// Send mining information to all connections
	for _, v := range m.allconns {
		stuffbts := m.penddingBlockMsg.Serialize()
		message.MsgSendToTcpConn(v.conn, message.MinerWorkMsgTypeMiningBlock, stuffbts)
	}

	// ok
}

func (m *MinerServer) StopMining() {

}
