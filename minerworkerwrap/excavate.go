package minerworkerwrap

import (
	"github.com/hacash/core/interfacev2"
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
func (g *WorkerWrap) Excavate(miningStuffCh chan interfacev2.PowWorkerMiningStuffItem, resultCh chan interfacev2.PowWorkerMiningStuffItem) {

	g.miningStuffCh = miningStuffCh
	g.resultCh = resultCh
}
