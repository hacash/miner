package miningpool

import (
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/miner/message"
	"net"
	"sync"
)

type RealtimePeriod struct {
	minerpool *MinerPool

	targetBlock      interfaces.Block
	realtimeAccounts map[string]*Account // [*Account]

	autoIncrementCoinbaseMsgNum uint32

	outputBlockCh *chan interfaces.Block

	changeLock sync.Mutex
}

func NewRealtimePeriod(minerpool *MinerPool, block interfaces.Block) *RealtimePeriod {
	per := &RealtimePeriod{
		minerpool:                   minerpool,
		targetBlock:                 block,
		realtimeAccounts:            make(map[string]*Account),
		autoIncrementCoinbaseMsgNum: 0,
		outputBlockCh:               nil,
	}
	return per
}

func (r *RealtimePeriod) getAutoIncrementCoinbaseMsgNum() uint32 {
	r.changeLock.Lock()
	defer r.changeLock.Unlock()

	r.autoIncrementCoinbaseMsgNum += 1
	return r.autoIncrementCoinbaseMsgNum
}

func (r *RealtimePeriod) sendMiningStuffMsg(conn net.Conn) {
	msgobj := message.NewPowMasterMsg()
	msgobj.CoinbaseMsgNum = fields.VarInt4(r.getAutoIncrementCoinbaseMsgNum())
	// send data
	data, _ := msgobj.Serialize()
	conn.Write(data)
}

// find ok
func (r *RealtimePeriod) successFindNewBlock(block interfaces.Block) {
	if r.outputBlockCh != nil {
		*r.outputBlockCh <- block // 挖出区块，传递给miner
	}
}

// 结束当前挖矿
func (r *RealtimePeriod) endCurrentMining() {
	for _, acc := range r.realtimeAccounts {
		clients := acc.activeClients.ToSlice()
		for _, cli := range clients {
			client := cli.(*Client)
			client.conn.Write([]byte("end_current_mining"))
		}
	}
}
