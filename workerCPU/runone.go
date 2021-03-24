package workerCPU

import (
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/mint/difficulty"
	"github.com/hacash/x16rs"
)

type miningBlockReturn struct {
	stopKind      byte
	isSuccess     bool
	nonceBytes    []byte
	powerHash     []byte
	blockHeadMeta interfaces.Block
}

func (c *CPUWorker) runOne(stuffitem interfaces.PowWorkerMiningStuffItem, stopMark *byte) *miningBlockReturn {
	newblockheadmeta := stuffitem.GetHeadMetaBlock()

	//fmt.Println("CPUWorker.runOne() mrkl:", newblockheadmeta.GetMrklRoot().ToHex())

	startNonce := uint32(0)
	endNonce := uint32(4294967295)
	//endNonce := uint32(429496)

	workStuff := blocks.CalculateBlockHashBaseStuff(newblockheadmeta)
	targethashdiff := difficulty.Uint32ToHash(newblockheadmeta.GetHeight(), newblockheadmeta.GetDifficulty())
	// run
	//fmt.Println( "targethashdiff:", hex.EncodeToString(targethashdiff) )
	// ========= test start =========
	//time.Sleep(time.Second)
	// ========= test end   =========
	stopkind, issuccess, noncebytes, powerhash := x16rs.MinerNonceHashX16RS(newblockheadmeta.GetHeight(), c.config.IsReportPower, stopMark, startNonce, endNonce, targethashdiff, workStuff)
	//fmt.Println("x16rs.MinerNonceHashX16RS finish ", issuccess,  binary.LittleEndian.Uint32(noncebytes[0:4]), startNonce, endNonce)
	if issuccess {
		// return success block
		*stopMark = 1 // stop all others
		//fmt.Println("start c.successBlockCh <- newblock")
		return &miningBlockReturn{
			stopkind,
			true,
			noncebytes,
			powerhash,
			newblockheadmeta,
		}
		//fmt.Println("end ... c.successBlockCh <- newblock")
	} else if c.config.IsReportPower {
		return &miningBlockReturn{
			stopkind,
			false,
			noncebytes,
			powerhash,
			newblockheadmeta,
		}
	}
	return nil
}
