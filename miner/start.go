package miner

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/sys"
	"github.com/hacash/mint"
	"github.com/hacash/mint/coinbase"
	"github.com/hacash/mint/difficulty"
	"strings"
	"sync/atomic"
	"time"
)

func (m *Miner) doStartMining() {

	defer func() {
		// set mining status stop
		atomic.StoreUint32(m.isMiningStatus, 0)
	}()

	// start mining
	last, _, err := m.blockchain.GetChainEngineKernel().LatestBlock()
	if err != nil {
		panic(err)
	}
	// pick up txs from pool
	pikuptxs := m.txpool.CopyTxsOrderByFeePurity(last.GetHeight()+1, mint.SingleBlockMaxTxCount, mint.SingleBlockMaxSize*2)
	//fmt.Println("doStartMining pikuptxs", pikuptxs)
	// create next block
	pikuptxsList := make([]interfaces.Transaction, len(pikuptxs))
	for i, v := range pikuptxs {
		pikuptxsList[i] = v.(interfaces.Transaction)
	}
	nextblock, removetxs, totaltxsize, e1 := m.blockchain.CreateNextBlockByValidateTxs(pikuptxsList)
	if e1 != nil {
		panic(e1)
	}
	m.txpool.RemoveTxsOnNextBlockArrive(removetxs) // remove
	// update set coinbase reward address
	coinbase.UpdateBlockCoinbaseMessage(nextblock, m.config.CoinbaseMessage)
	coinbase.UpdateBlockCoinbaseAddress(nextblock, m.config.Rewards)

	// update mkrl root
	nextblock.SetMrklRoot(blocks.CalculateMrklRoot(nextblock.GetTrsList()))

	nextblockHeight := nextblock.GetHeight()

	if nextblockHeight%mint.AdjustTargetDifficultyNumberOfBlocks == 0 {
		diff1 := last.GetDifficulty()
		diff2 := nextblock.GetDifficulty()
		tarhx1 := hex.EncodeToString(difficulty.Uint32ToHash(last.GetHeight(), diff1))
		tarhx2 := hex.EncodeToString(difficulty.Uint32ToHash(nextblockHeight, diff2))
		costtime, err := difficulty.ReadPrev288BlockTimestamp(m.blockchain.GetChainEngineKernel().StateRead().BlockStoreRead(), nextblockHeight)
		if err == nil {
			costtime = nextblock.GetTimestamp() - costtime
		}
		targettime := mint.AdjustTargetDifficultyNumberOfBlocks * mint.EachBlockRequiredTargetTime
		fmt.Printf("\n== target difficulty change == %d == -> == (%ds/%ds) == %d -> %d == %s -> %s \n\n",
			nextblockHeight,
			costtime, targettime,
			diff1, diff2,
			strings.TrimRight(tarhx1, "0"),
			strings.TrimRight(tarhx2, "0"),
		)
	}

	fmt.Printf("do mining... block height: %d, txs: %d, prev: %s..., difficulty: %s, size: %fkb, time: %s\n",
		nextblockHeight,
		nextblock.GetCustomerTransactionCount(),
		string([]byte(nextblock.GetPrevHash().ToHex())[0:32]),
		strings.TrimRight(hex.EncodeToString(difficulty.Uint32ToHash(nextblockHeight, nextblock.GetDifficulty())), "0"),
		float64(totaltxsize)/1024,
		time.Unix(int64(nextblock.GetTimestamp()), 0).Format("01/02 15:04:05"),
	)

	if sys.TestDebugLocalDevelopmentMark || sys.NotCheckBlockDifficultyForMiner {
		// Mining sleep time during development test
		time.Sleep(time.Second * (time.Duration(mint.EachBlockRequiredTargetTime + 1)))
	}

	//fmt.Println("m.powmaster.Excavate(nextblock, backBlockCh) MrklRoot:", nextblock.GetMrklRoot().ToHex())
	// excavate block
	backBlockCh := make(chan interfaces.Block, 1)
	m.powmaster.DoMining(nextblock, backBlockCh)

	//fmt.Println("finifsh m.powmaster.Excavate nextblock")

	var miningSuccessBlock interfaces.Block = nil
	select {
	case miningSuccessBlock = <-backBlockCh:
	case <-m.stopSignCh:
		// fmt.Println("return <- m.stopSignCh:")
		return // stop mining
	}
	// mark stop
	atomic.StoreUint32(m.isMiningStatus, 0)
	//fmt.Println("select miningSuccessBlock ok", miningSuccessBlock)
	/*
		dddd, _ := miningSuccessBlock.Serialize()
		fmt.Println(dddd)
		fmt.Println(miningSuccessBlock.GetTransactions()[0].Serialize())
		miningSuccessBlock, _, _ = blocks.ParseBlock(dddd, 0)
		fmt.Println(miningSuccessBlock.GetTransactions()[0].Serialize())
	*/
	if nextblockHeight < 288*100 {
		time.Sleep(time.Second)
	}
	// mining success
	if miningSuccessBlock != nil {
		inserterr := m.blockchain.GetChainEngineKernel().InsertBlock(miningSuccessBlock, "mining")
		if inserterr == nil {
			coinbaseStr := ""
			coinbasetx := miningSuccessBlock.GetTrsList()[0]
			coinbaseStr += coinbasetx.GetAddress().ToReadable()
			coinbaseStr += " + " + coinbase.BlockCoinBaseReward(miningSuccessBlock.GetHeight()).ToFinString()
			// show success
			fmt.Printf("[⬤◆◆] Successfully mined a block height: %d, txs: %d, hash: %s, coinbase: %s, time: %s \n",
				miningSuccessBlock.GetHeight(),
				miningSuccessBlock.GetCustomerTransactionCount(),
				miningSuccessBlock.Hash().ToHex(),
				coinbaseStr,
				time.Now().Format("01/02 15:04:05"),
			)
		} else {
			fmt.Println("[Miner Error]", inserterr.Error())
			//m.StartMining()
		}
	}
}
