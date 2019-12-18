package miningpool

import (
	"encoding/binary"
	"github.com/hacash/core/blocks"
	"github.com/hacash/miner/message"
	"github.com/hacash/mint/coinbase"
	"time"
)

func (a *Account) successFindNewBlock(msg *message.PowMasterMsg) {
	minerpool := a.realtimePeriod.minerpool

	// cache
	minerpool.prevRealtimePeriod = minerpool.currentRealtimePeriod

	// copy data
	copyblock := a.workBlock.CopyForMining()
	// update
	coinbase.UpdateBlockCoinbaseMessageForMiner(copyblock, uint32(msg.CoinbaseMsgNum))
	copyblock.SetNonce(binary.BigEndian.Uint32(msg.NonceBytes))
	copyblock.SetMrklRoot(blocks.CalculateMrklRoot(copyblock.GetTransactions()))
	copyblock.SetOriginMark("mining") // set origin
	copyblock.Fresh()
	// mark success account
	a.miningSuccessBlockHash = copyblock.Hash()
	// create next period
	minerpool.currentRealtimePeriod = NewRealtimePeriod(minerpool, a.realtimePeriod.targetBlock)
	// insert new block
	a.realtimePeriod.successFindNewBlock(copyblock)

	go func() {
		<-time.Tick(time.Second * 33) // 33 秒后去结算 period
		minerpool.settleOnePeriod(a.realtimePeriod)
	}()
}
