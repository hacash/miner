package workerGPU

import (
	"github.com/hacash/core/interfaces"
	"github.com/hacash/miner/message"
)

func (l *LocalGpuWorker) SetCoinbaseMsgNum(coinbaseMsgNum uint32) {
	l.coinbaseMsgNum = coinbaseMsgNum
}

// stop mining
func (l *LocalGpuWorker) StopMining() {
	l.stopMarks.Range(func(k interface{}, v interface{}) bool {
		mk := v.(*byte)
		*mk = 1 // set stop
		return false
	})
}

// do mining
func (l *LocalGpuWorker) Excavate(inputblockheadmeta interfaces.Block, outputCh chan message.PowMasterMsg) {

	l.stepLock.Lock()
	defer l.stepLock.Unlock()

	l.StopMining() // stop old all

}
