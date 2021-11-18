package localcpu

import (
	"sync"
)

type LocalCPUPowMaster struct {
	config *LocalCPUPowMasterConfig

	coinbaseMsgNum uint32

	//currentWorkers mapset.Set
	stopMarks sync.Map

	stepLock sync.RWMutex
}

func NewLocalCPUPowMaster(cnf *LocalCPUPowMasterConfig) *LocalCPUPowMaster {

	miner := &LocalCPUPowMaster{
		config: cnf,
	}

	return miner
}

func (l *LocalCPUPowMaster) SetCoinbaseMsgNum(coinbaseMsgNum uint32) {
	l.stepLock.Lock()
	l.coinbaseMsgNum = coinbaseMsgNum
	l.stepLock.Unlock()
}
