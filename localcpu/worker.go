package localcpu

import (
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/mint/difficulty"
	"github.com/hacash/x16rs"
	"sync/atomic"
)

type miningBlockReturn struct {
	stopKind       byte
	isSuccess      bool
	coinbaseMsgNum uint32
	nonceBytes     []byte
	powerHash      []byte
	blockHeadMeta  interfaces.Block
}

type CPUWorker struct {
	returnPowerHash bool

	stopMark *byte

	coinbaseMsgNum uint32

	successMiningMark *uint32

	successBlockCh chan miningBlockReturn
}

func NewCPUWorker(successMiningMark *uint32, successBlockCh chan miningBlockReturn, coinbaseMsgNum uint32, stopMark *byte) *CPUWorker {
	worker := &CPUWorker{
		returnPowerHash:   false,
		successMiningMark: successMiningMark,
		successBlockCh:    successBlockCh,
		coinbaseMsgNum:    coinbaseMsgNum,
		stopMark:          stopMark,
	}
	return worker
}

func (c *CPUWorker) RunMining(newblockheadmeta interfaces.Block, startNonce uint32, endNonce uint32) bool {
	workStuff := blocks.CalculateBlockHashBaseStuff(newblockheadmeta)
	targethashdiff := difficulty.Uint32ToHash(newblockheadmeta.GetHeight(), newblockheadmeta.GetDifficulty())
	// run
	//fmt.Println( "targethashdiff:", hex.EncodeToString(targethashdiff) )
	// ========= test start =========
	//time.Sleep(time.Second)
	// ========= test end   =========
	stopkind, issuccess, noncebytes, powerhash := x16rs.MinerNonceHashX16RS(newblockheadmeta.GetHeight(), c.returnPowerHash, c.stopMark, startNonce, endNonce, targethashdiff, workStuff)
	//fmt.Println("x16rs.MinerNonceHashX16RS finish ", issuccess,  binary.LittleEndian.Uint32(noncebytes[0:4]), startNonce, endNonce)
	if issuccess && atomic.CompareAndSwapUint32(c.successMiningMark, 0, 1) {
		// return success block
		*c.stopMark = 1 // set stop mark for all cpu worker
		//fmt.Println("start c.successBlockCh <- newblock")
		c.successBlockCh <- miningBlockReturn{
			stopkind,
			true,
			c.coinbaseMsgNum,
			noncebytes,
			nil,
			newblockheadmeta,
		}
		//fmt.Println("end ... c.successBlockCh <- newblock")
		return true
	} else if c.returnPowerHash {
		c.successBlockCh <- miningBlockReturn{
			stopkind,
			false,
			c.coinbaseMsgNum,
			noncebytes,
			powerhash,
			newblockheadmeta,
		}
		return false
	}
	return false
}
