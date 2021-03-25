package workerGPU

import (
	"bytes"
	"github.com/hacash/x16rs/opencl/execute"
	"sync"
)

type GpuWorker struct {
	config   *GpuWorkerConfig
	gpuMiner *execute.GpuMiner

	//currentWorkers mapset.Set
	stopMarks sync.Map
	stepLock  sync.Mutex
}

func NewGpuWorker(config *GpuWorkerConfig) *GpuWorker {
	return &GpuWorker{
		config: config,
	}
}

// 关闭算力统计
func (g *GpuWorker) CloseUploadPower() {

}

// 开始采矿
func (mr *GpuWorker) DoMining(blockHeight uint64, retmaxhash bool, stopmark *byte, hashstart uint32, hashend uint32, tarhashvalue []byte, blockheadmeta []byte) (byte, bool, []byte, []byte) {

	return 0, false, []byte{0, 0, 0, 0}, bytes.Repeat([]byte{255}, 32)
}
