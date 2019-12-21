package minerpool

import (
	"github.com/hacash/core/interfaces"
)

// stop all
func (p *MinerPool) StopMining() {
	// do nothing
}

func (p *MinerPool) Excavate(inputBlock interfaces.Block, outputBlockCh chan interfaces.Block) {
	p.periodChange.Lock()
	defer p.periodChange.Unlock()

	if p.currentRealtimePeriod == nil {
		p.currentRealtimePeriod = NewRealtimePeriod(p, inputBlock)
	} else {
		// cache
		p.prevRealtimePeriod = p.currentRealtimePeriod
		// create next period
		p.currentRealtimePeriod = NewRealtimePeriod(p, inputBlock)
	}

	// 设置新的挖矿区块，以供客户端请求
	p.currentRealtimePeriod.targetBlock = inputBlock
	p.currentRealtimePeriod.outputBlockCh = &outputBlockCh

	// 结束当前的全部挖矿
	if p.prevRealtimePeriod != nil {
		p.prevRealtimePeriod.endCurrentMining()
	}
	//p.currentRealtimePeriod.endCurrentMining()
}
