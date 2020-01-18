package minerpool

import (
	"github.com/hacash/core/interfaces"
	"sync/atomic"
	"time"
)

// stop all
func (p *MinerPool) StopMining() {
	// do nothing
}

func (p *MinerPool) Excavate(inputBlock interfaces.Block, outputBlockCh chan interfaces.Block) {
	p.periodChange.Lock()
	defer p.periodChange.Unlock()
	defer func() {
		p.successFindNewBlockHashOnce = nil // reset status
	}()

	prevblockMiningSuccess := p.successFindNewBlockHashOnce != nil &&
		p.successFindNewBlockHashOnce.Equal( inputBlock.GetPrevHash() )

	var endPrev bool = false
	var sendCurrentRestartMining bool = false

	if p.currentRealtimePeriod == nil {
		p.currentRealtimePeriod = NewRealtimePeriod(p, inputBlock)
	} else {
		if prevblockMiningSuccess {
			// cache
			p.prevRealtimePeriod = p.currentRealtimePeriod
			// create next period
			p.currentRealtimePeriod = NewRealtimePeriod(p, inputBlock)
			// end current
			endPrev = true
		}else{
			sendCurrentRestartMining = true
		}
	}

	// 设置新的挖矿区块，以供客户端请求
	atomic.StoreUint32( &p.currentRealtimePeriod.autoIncrementCoinbaseMsgNum, 0 )
	p.currentRealtimePeriod.targetBlock = inputBlock
	p.currentRealtimePeriod.outputBlockCh = &outputBlockCh

	if endPrev {
		p.prevRealtimePeriod.endCurrentMining()
	}

	if sendCurrentRestartMining {
		p.currentRealtimePeriod.sendMiningStuffMsgToAllClient()
	}

	// 结束当前的全部挖矿
	if p.prevRealtimePeriod != nil {
		//p.prevRealtimePeriod.endCurrentMining()
		// 确认应得的奖励，并开始打币流程
		if prevblockMiningSuccess {
			go func() {
				time.Sleep(time.Second * 1)
				p.confirmRewards(inputBlock.GetHeight(), p.prevRealtimePeriod)
				time.Sleep(time.Second * 1)
				p.startDoTransfer(inputBlock.GetHeight(), p.prevRealtimePeriod)
			}()
		}
	}
}
