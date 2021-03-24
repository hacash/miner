package workerCPU

import (
	"github.com/hacash/core/interfaces"
	"sync"
)

type CPUWorker struct {
	config *CPUWorkerConfig

	pendingBlockHeight uint64
	nextMarks          sync.Map

	successMiningMark *uint32

	miningStuffCh chan interfaces.PowWorkerMiningStuffItem
	resultCh      chan interfaces.PowWorkerMiningStuffItem

	// 挖矿流程锁
	miningstreamlock sync.Mutex
}

func NewCPUWorker(cnf *CPUWorkerConfig) *CPUWorker {

	miner := &CPUWorker{
		config: cnf,
	}

	return miner
}

// 初始化
func (c *CPUWorker) InitStart() error {
	return nil
}

// 关闭统计算力
func (c *CPUWorker) CloseUploadPower() {
	//c.config.IsReportPower = false
}

// stop mining
func (l *CPUWorker) NextMiningOld(nextheight uint64) {
	if l.pendingBlockHeight == nextheight {
		return // 挖掘高度相同，忽略
	}
	var smm uint32 = 0
	l.successMiningMark = &smm // 未成功标记
	l.pendingBlockHeight = nextheight
	l.nextMarks.Range(func(k interface{}, v interface{}) bool {
		mk := v.(*byte)
		*mk = 1 // set stop
		return false
	})
}
