package localcpu

import (
	"sync"
)

type LocalCPUPowMaster struct {
	config *LocalCPUPowMasterConfig

	//currentWorkers mapset.Set
	stopMarks sync.Map

	stepLock sync.Mutex
}

func NewLocalCPUPowMaster(cnf *LocalCPUPowMasterConfig) *LocalCPUPowMaster {

	miner := &LocalCPUPowMaster{
		config: cnf,
		//currentWorkers: mapset.NewSet(),
	}
	return miner
}
