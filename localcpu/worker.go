package localcpu

import (
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/mint/difficulty"
	"github.com/hacash/x16rs"
	"sync/atomic"
	"time"
)

type successBlockReturn struct {
	coinbaseMsgNum uint32
	nonceBytes     []byte
	blockHeadMeta  interfaces.Block
}

type CPUWorker struct {
	stopMark *byte

	coinbaseMsgNum uint32

	successMiningMark *uint32

	successBlockCh chan successBlockReturn
}

func NewCPUWorker(successMiningMark *uint32, successBlockCh chan successBlockReturn, coinbaseMsgNum uint32, stopMark *byte) *CPUWorker {
	worker := &CPUWorker{
		successMiningMark: successMiningMark,
		successBlockCh:    successBlockCh,
		coinbaseMsgNum:    coinbaseMsgNum,
		stopMark:          stopMark,
	}
	return worker
}

func (c *CPUWorker) RunMining(newblockheadmeta interfaces.Block, startNonce uint32, endNonce uint32) bool {
	loopnum := int(newblockheadmeta.GetHeight()/50000) + 1
	if loopnum > 16 {
		loopnum = 16
	}
	workStuff := blocks.CalculateBlockHashBaseStuff(newblockheadmeta)
	targethashdiff := difficulty.Uint32ToHash(newblockheadmeta.GetHeight(), newblockheadmeta.GetDifficulty())
	// run
	//fmt.Println( "targethashdiff:", hex.EncodeToString(targethashdiff) )
	// ========= test start =========
	time.Sleep(time.Second)
	// ========= test end   =========
	issuccess, noncebytes, _ := x16rs.MinerNonceHashX16RS(loopnum, false, c.stopMark, startNonce, endNonce, targethashdiff, workStuff)
	//fmt.Println("x16rs.MinerNonceHashX16RS finish")
	if issuccess && atomic.CompareAndSwapUint32(c.successMiningMark, 0, 1) {
		// return success block
		//fmt.Println("start c.successBlockCh <- newblock")
		c.successBlockCh <- successBlockReturn{
			c.coinbaseMsgNum,
			noncebytes,
			newblockheadmeta,
		}
		//fmt.Println("end ... c.successBlockCh <- newblock")
		return true
	}
	return false
}
