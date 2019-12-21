package minerpool

import (
	"encoding/binary"
	"github.com/hacash/core/blocks"
	"github.com/hacash/miner/message"
	"github.com/hacash/mint/coinbase"
	"time"
)

func (a *Account) successFindNewBlock(msg *message.PowMasterMsg) {
	minerpool := a.realtimePeriod.minerpool

	// copy data
	copyblock := a.workBlock.CopyForMining()

	//fmt.Println("========================================")
	//fmt.Println(msg.BlockHeadMeta)
	//fmt.Println("--------")
	//fmt.Println(a.workBlock)
	//fmt.Println("--------")
	//fmt.Println(copyblock)

	// update
	coinbase.UpdateBlockCoinbaseMessageForMiner(copyblock, uint32(msg.CoinbaseMsgNum))
	copyblock.SetNonce(binary.BigEndian.Uint32(msg.NonceBytes))
	copyblock.SetMrklRoot(blocks.CalculateMrklRoot(copyblock.GetTransactions()))
	copyblock.SetOriginMark("mining") // set origin
	copyblock.Fresh()
	//fmt.Println("--------")
	//fmt.Println(copyblock)
	//fmt.Println("========================================")
	// mark success account
	a.miningSuccessBlock = copyblock
	a.realtimePeriod.miningSuccessBlock = copyblock
	// insert new block
	a.realtimePeriod.successFindNewBlock(copyblock)
	// store success
	minerpool.saveFoundBlockHash(copyblock.GetHeight(), copyblock.Hash())
	// settle 结算
	go func() {
		<-time.Tick(time.Second * 5) // 33 秒后去结算 period
		minerpool.settleOnePeriod(a.realtimePeriod)
	}()
}
