package minerpool

import (
	"encoding/binary"
	"github.com/hacash/core/blocks"
	"github.com/hacash/miner/message"
	"github.com/hacash/mint/coinbase"
)

func (a *Account) successFindNewBlock(msg *message.PowMasterMsg) {
	minerpool := a.realtimePeriod.minerpool

	//minerpool.periodChange.Lock()
	//defer minerpool.periodChange.Unlock()

	// copy data
	copyblock := a.realtimePeriod.targetBlock.CopyForMiningV3()

	//fmt.Println("========================================")
	//fmt.Println(msg.BlockHeadMeta)
	//fmt.Println("--------")
	//fmt.Println(a.workBlock)
	//fmt.Println("--------")
	//fmt.Println(copyblock)

	// update
	coinbase.UpdateBlockCoinbaseMessageForMiner(copyblock, uint32(msg.CoinbaseMsgNum))
	copyblock.SetNonce(binary.BigEndian.Uint32(msg.NonceBytes))
	copyblock.SetMrklRoot(blocks.CalculateMrklRoot(copyblock.GetTrsList()))
	copyblock.SetOriginMark("mining") // set origin
	copyblock.Fresh()

	// mark new Period
	minerpool.successFindNewBlockHashOnce = copyblock.Hash()
	//fmt.Println("--------")
	//fmt.Println(copyblock)
	//fmt.Println("========================================")
	// mark success account
	a.miningSuccessBlock = copyblock
	a.realtimePeriod.miningSuccessBlock = copyblock
	// insert new block
	//fmt.Println("a.realtimePeriod.successFindNewBlock MrklRoot:", copyblock.GetMrklRoot().ToHex())
	a.realtimePeriod.successFindNewBlock(copyblock)
	// settle 结算
	minerpool.createSettlementPeriod(a, a.realtimePeriod, copyblock)
	/*
		go func() {
			<-time.Tick(time.Second * 33) // 33 秒后去结算 period
			minerpool.settleRealtimePeriodCh <- a.realtimePeriod
		}()
	*/
}
