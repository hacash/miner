package localcpu

import (
	"sync"
)

type LocalCPUPowMaster struct {
	config *LocalCPUPowMasterConfig

	coinbaseMsgNum uint32

	//currentWorkers mapset.Set
	stopMarks sync.Map

	stepLock sync.Mutex
}

func NewLocalCPUPowMaster(cnf *LocalCPUPowMasterConfig) *LocalCPUPowMaster {

	miner := &LocalCPUPowMaster{
		config: cnf,
	}

	return miner
}

func (l *LocalCPUPowMaster) SetCoinbaseMsgNum(coinbaseMsgNum uint32) {
	l.coinbaseMsgNum = coinbaseMsgNum
}
