package localcpu

import (
	"encoding/binary"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/mint/coinbase"
	"github.com/hacash/mint/difficulty"
	"github.com/hacash/x16rs"
	"sync/atomic"
	"time"
)

type CPUWorker struct {
	stopMark *byte

	coinbaseMsgNum uint32

	successMiningMark *uint32

	successBlockCh chan interfaces.Block
}

func NewCPUWorker(successMiningMark *uint32, successBlockCh chan interfaces.Block, coinbaseMsgNum uint32, stopMark *byte) *CPUWorker {
	worker := &CPUWorker{
		successMiningMark: successMiningMark,
		successBlockCh:    successBlockCh,
		coinbaseMsgNum:    coinbaseMsgNum,
		stopMark:          stopMark,
	}
	return worker
}

func (c *CPUWorker) RunMining(newblock interfaces.Block, startNonce uint32, endNonce uint32) bool {
	loopnum := int(newblock.GetHeight()/50000) + 1
	if loopnum > 16 {
		loopnum = 16
	}
	workStuff := blocks.CalculateBlockHashBaseStuff(newblock)
	targethashdiff := difficulty.Uint32ToHash(newblock.GetHeight(), newblock.GetDifficulty())
	// run
	//fmt.Println( "targethashdiff:", hex.EncodeToString(targethashdiff) )
	// ========= test start =========
	time.Sleep(time.Second)
	// ========= test end   =========
	issuccess, noncebytes, _ := x16rs.MinerNonceHashX16RS(loopnum, false, c.stopMark, startNonce, endNonce, targethashdiff, workStuff)
	//fmt.Println("x16rs.MinerNonceHashX16RS finish")
	if issuccess && atomic.CompareAndSwapUint32(c.successMiningMark, 0, 1) {
		if c.coinbaseMsgNum > 0 {
			coinbase.UpdateBlockCoinbaseMessageForMiner(newblock, c.coinbaseMsgNum)
		}
		newblock.SetNonce(binary.BigEndian.Uint32(noncebytes))
		newblock.SetMrklRoot(blocks.CalculateMrklRoot(newblock.GetTransactions())) // update mrkl
		newblock.Fresh()
		// return success block
		//fmt.Println("start c.successBlockCh <- newblock")
		c.successBlockCh <- newblock
		//fmt.Println("end ... c.successBlockCh <- newblock")
		return true
	}
	return false
}
