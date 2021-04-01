package minerworkerwrap

import (
	"github.com/hacash/core/interfaces"
	"sync"
)

type WorkerWrap struct {
	config    *WorkerWrapConfig
	powdevice interfaces.PowDevice

	//currentWorkers mapset.Set
	stopMarks sync.Map
	stepLock  sync.Mutex

	// chan
	miningStuffCh chan interfaces.PowWorkerMiningStuffItem
	resultCh      chan interfaces.PowWorkerMiningStuffItem
}

func NewWorkerWrap(config *WorkerWrapConfig) *WorkerWrap {
	return &WorkerWrap{
		config: config,
	}
}

// 关闭算力统计
func (g *WorkerWrap) CloseUploadPower() {
	g.powdevice.CloseUploadPower()
}
