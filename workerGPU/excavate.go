package workerGPU

import (
	"github.com/hacash/core/interfaces"
)

// stop mining
func (g *GpuWorker) StopAllMining() {
	g.stopMarks.Range(func(k interface{}, v interface{}) bool {
		mk := v.(*byte)
		*mk = 1 // set stop
		return false
	})
}

// do mining
func (g *GpuWorker) Excavate(miningStuffCh chan interfaces.PowWorkerMiningStuffItem, resultCh chan interfaces.PowWorkerMiningStuffItem) {

	g.StopAllMining() // stop old all

}
