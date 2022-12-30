package localgpu

import (
	"github.com/hacash/core/interfaces"
	"sync"
)

type LocalGPUPowMaster struct {
	config *LocalGPUPowMasterConfig

	coinbaseMsgNum uint32

	//currentWorkers mapset.Set
	stopMarks sync.Map

	stepLock sync.RWMutex

	miningStuffCh chan interfaces.PowWorkerMiningStuffItem
	resultCh      chan interfaces.PowWorkerMiningStuffItem
}

func NewLocalGPUPowMaster(cnf *LocalGPUPowMasterConfig) *LocalGPUPowMaster {

	miner := &LocalGPUPowMaster{
		config: cnf,
	}

	return miner
}

func (l *LocalGPUPowMaster) SetCoinbaseMsgNum(coinbaseMsgNum uint32) {
	l.stepLock.Lock()
	l.coinbaseMsgNum = coinbaseMsgNum
	l.stepLock.Unlock()
}
