package miner

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
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
	last, err := m.blockchain.State().ReadLastestBlockHeadAndMeta()
	if err != nil {
		panic(err)
	}
	// pick up txs from pool
	pikuptxs := m.txpool.CopyTxsOrderByFeePurity(last.GetHeight()+1, 2000, mint.SingleBlockMaxSize*2)
	// create next block
	nextblock, removetxs, totaltxsize, e1 := m.blockchain.CreateNextBlockByValidateTxs(pikuptxs)
	if e1 != nil {
		panic(e1)
	}
	m.txpool.RemoveTxsOnNextBlockArrive(removetxs) // remove
	// update set coinbase reward address
	coinbase.UpdateBlockCoinbaseMessage(nextblock, m.config.CoinbaseMessage)
	coinbase.UpdateBlockCoinbaseAddress(nextblock, m.config.Rewards)

	// update mkrl root
	nextblock.SetMrklRoot(blocks.CalculateMrklRoot(nextblock.GetTransactions()))

	nextblockHeight := nextblock.GetHeight()

	if nextblockHeight%mint.AdjustTargetDifficultyNumberOfBlocks == 0 {
		diff1 := last.GetDifficulty()
		diff2 := nextblock.GetDifficulty()
		tarhx1 := hex.EncodeToString(difficulty.Uint32ToHash(last.GetHeight(), diff1))
		tarhx2 := hex.EncodeToString(difficulty.Uint32ToHash(nextblockHeight, diff2))
		costtime, err := m.blockchain.ReadPrev288BlockTimestamp(nextblockHeight)
		if err == nil {
			costtime = nextblock.GetTimestamp() - costtime
		}
		targettime := mint.AdjustTargetDifficultyNumberOfBlocks * mint.EachBlockRequiredTargetTime
		fmt.Printf("\n== target difficulty change == %d == -> == (%ds/%ds) == %d -> %d == %s -> %s \n\n",
			nextblockHeight,
			costtime, targettime,
			diff1, diff2,
			strings.TrimRight(string([]byte(tarhx1)[0:32]), "0"),
			strings.TrimRight(string([]byte(tarhx2)[0:32]), "0"),
		)
	}

	fmt.Printf("do mining... block height: %d, txs: %d, prev: %s..., difficulty: %d, size: %fkb, time: %s\n",
		nextblockHeight,
		nextblock.GetTransactionCount()-1,
		string([]byte(nextblock.GetPrevHash().ToHex())[0:32]),
		nextblock.GetDifficulty(),
		float64(totaltxsize)/1024,
		time.Unix(int64(nextblock.GetTimestamp()), 0).Format("01/02 15:04:05"),
	)

	//fmt.Println("m.powserver.Excavate(nextblock, backBlockCh) MrklRoot:", nextblock.GetMrklRoot().ToHex())
	// excavate block
	backBlockCh := make(chan interfaces.Block, 1)
	m.powserver.Excavate(nextblock, backBlockCh)

	//fmt.Println("finifsh m.powserver.Excavate nextblock")

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
	// mining success
	if miningSuccessBlock != nil {
		miningSuccessBlock.SetOriginMark("mining")
		inserterr := m.blockchain.InsertBlock(miningSuccessBlock)
		if inserterr == nil {
			coinbaseStr := ""
			coinbasetx := miningSuccessBlock.GetTransactions()[0]
			coinbaseStr += coinbasetx.GetAddress().ToReadable()
			coinbaseStr += " + " + coinbase.BlockCoinBaseReward(miningSuccessBlock.GetHeight()).ToFinString()
			// show success
			fmt.Printf("⬤ mining new block height: %d, txs: %d, hash: %s, coinbase: %s, successfully!\n",
				miningSuccessBlock.GetHeight(),
				miningSuccessBlock.GetTransactionCount()-1,
				miningSuccessBlock.Hash().ToHex(),
				coinbaseStr,
			)
		} else {
			fmt.Println("[Miner Error]", inserterr.Error())
			m.StartMining()
		}
	}
}
