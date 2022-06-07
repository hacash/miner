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

// Turn off force statistics
func (g *WorkerWrap) CloseUploadHashrate() {
	g.powdevice.CloseUploadHashrate()
}
