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

	//defer func() {
	//	p.successFindNewBlockHashOnce = nil // reset status
	//}()

	//prevblockMiningSuccess := p.successFindNewBlockHashs.Contains(string(inputBlock.GetPrevHash()))

	//var endPrev bool = false
	//var sendCurrentRestartMining bool = false

	if p.currentRealtimePeriod == nil {
		p.currentRealtimePeriod = NewRealtimePeriod(p, inputBlock)
	} /*else {
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
	}*/

	// Set up new mining blocks for client requests
	atomic.StoreUint32(&p.currentRealtimePeriod.autoIncrementCoinbaseMsgNum, 0)
	p.currentRealtimePeriod.targetBlock = inputBlock
	p.currentRealtimePeriod.outputBlockCh = &outputBlockCh

	//if endPrev {
	//	p.prevRealtimePeriod.endCurrentMining()
	//}

	//if sendCurrentRestartMining {
	p.currentRealtimePeriod.sendMiningStuffMsgToAllClient()
	//}

	// End all current mining
	if p.prevRealtimePeriod != nil {
		p.prevRealtimePeriod.endCurrentMining()
		// Confirm the deserved reward and start the coining process
		if p.prevRealtimePeriod != nil {
			go func() {
				time.Sleep(time.Second * 1)
				p.confirmRewards(inputBlock.GetHeight(), p.prevRealtimePeriod)
				time.Sleep(time.Second * 1)
				p.startDoTransfer(inputBlock.GetHeight(), p.prevRealtimePeriod)
			}()
			//}
		}

	}

	// mapset max size
	for {
		if p.successFindNewBlockHashs.Cardinality() > 32 {
			p.successFindNewBlockHashs.Pop()
		} else {
			break
		}
	}

}
