package minerworkerwrap

import (
	"github.com/hacash/core/interfaces"
)

// stop mining
func (g *WorkerWrap) StopAllMining() {
	g.stopMarks.Range(func(k interface{}, v interface{}) bool {
		mk := v.(*byte)
		*mk = 1 // set stop
		return true
	})
}

// do mining
func (g *WorkerWrap) Excavate(miningStuffCh chan interfaces.PowWorkerMiningStuffItem, resultCh chan interfaces.PowWorkerMiningStuffItem) {

	g.miningStuffCh = miningStuffCh
	g.resultCh = resultCh
}
