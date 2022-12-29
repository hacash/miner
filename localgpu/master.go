package localgpu

import (
	"sync"
)

type LocalGPUPowMaster struct {
	config *LocalGPUPowMasterConfig

	coinbaseMsgNum uint32

	//currentWorkers mapset.Set
	stopMarks sync.Map

	stepLock sync.RWMutex
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
