package minerpool

import (
	"github.com/hacash/core/interfaces"
	"time"
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
		if p.successFindNewBlockOnce == true {
			p.successFindNewBlockOnce = false // do once
			// cache
			p.prevRealtimePeriod = p.currentRealtimePeriod
			// create next period
			p.currentRealtimePeriod = NewRealtimePeriod(p, inputBlock)
		}
	}

	// 设置新的挖矿区块，以供客户端请求
	p.currentRealtimePeriod.targetBlock = inputBlock
	p.currentRealtimePeriod.outputBlockCh = &outputBlockCh

	// 结束当前的全部挖矿
	if p.prevRealtimePeriod != nil {
		p.prevRealtimePeriod.endCurrentMining()
		// 确认应得的奖励，并开始打币流程
		go func() {
			time.Sleep(time.Second)
			p.confirmRewards(inputBlock.GetHeight(), p.prevRealtimePeriod)
			time.Sleep(time.Millisecond * 150)
			p.startDoTransfer(inputBlock.GetHeight(), p.prevRealtimePeriod)
		}()

	}
	//p.currentRealtimePeriod.endCurrentMining()
}
