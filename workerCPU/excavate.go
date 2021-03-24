package workerCPU

import (
	"github.com/hacash/core/interfaces"
)

func (c *CPUWorker) Excavate(miningStuffCh chan interfaces.PowWorkerMiningStuffItem, resultCh chan interfaces.PowWorkerMiningStuffItem) {
	c.miningStuffCh = miningStuffCh
	c.resultCh = resultCh
}
