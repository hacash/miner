package minerworkerwrap

import (
	"github.com/hacash/core/interfacev2"
	"sync"
)

type WorkerWrap struct {
	config    *WorkerWrapConfig
	powdevice interfacev2.PowDevice

	//currentWorkers mapset.Set
	stopMarks sync.Map
	stepLock  sync.Mutex

	// chan
	miningStuffCh chan interfacev2.PowWorkerMiningStuffItem
	resultCh      chan interfacev2.PowWorkerMiningStuffItem
}

func NewWorkerWrap(config *WorkerWrapConfig) *WorkerWrap {
	return &WorkerWrap{
		config: config,
	}
}

// 关闭算力统计
func (g *WorkerWrap) CloseUploadHashrate() {
	g.powdevice.CloseUploadHashrate()
}
