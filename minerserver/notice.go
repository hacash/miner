package minerserver

import (
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/transactions"
	interfaces2 "github.com/hacash/miner/interfaces"
	"github.com/hacash/miner/message"
)

func (m *MinerServer) Init() error {
	return nil
}

func (m *MinerServer) DoMining(input interfaces.Block, resCh chan interfaces.Block) error {
	m.Excavate(input, resCh)
	return nil
}

// find block nonce or change coinbase message
// Issue
func (m *MinerServer) Excavate(input interfaces.Block, resCh chan interfaces.Block) {
	// Start mining new blocks
	m.changelock.Lock()
	defer m.changelock.Unlock()

	// Parsing mining messages
	var err error = nil
	var trslist = input.GetTrsList()
	if len(trslist) < 1 {
		return
	}
	cbtx, ok := trslist[0].(*transactions.Transaction_0_Coinbase)
	if !ok {
		return
	}
	mkrltree := blocks.PickMrklListForCoinbaseTxModify(trslist)
	var blkmsg = interfaces2.PoWStuffOverallData{
		BlockHeadMeta:     input,
		CoinbaseTx:        *cbtx.CopyForMining(),
		MrklCheckTreeList: mkrltree,
	}
	m.penddingBlockMsg = &blkmsg

	// Mining successful reporting channel
	m.successMintCh = resCh

	// fmt.Printf("send pending mining block stuff to %d worker of connected with server\n", len(m.allconns))

	// Send mining information to all connections
	stuffbts, err := m.penddingBlockMsg.Serialize()
	if err != nil {
		return
	}

	//fmt.Println(hex.EncodeToString(stuffbts))
	// send to all client
	for _, v := range m.allconns {
		message.MsgSendToTcpConn(v.conn, message.MinerWorkMsgTypeMiningBlock, stuffbts)
	}

	// ok
}

func (m *MinerServer) StopMining() {

}
