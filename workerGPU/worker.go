package workerGPU

import (
	"bytes"
	"github.com/hacash/x16rs/opencl/execute"
	"sync"
)

type LocalGpuWorker struct {
	gpuMiner *execute.GpuMiner

	// 数据
	coinbaseMsgNum uint32

	//currentWorkers mapset.Set
	stopMarks sync.Map
	stepLock  sync.Mutex
}

func NewGpuWorker() *LocalGpuWorker {
	return &LocalGpuWorker{}
}

// 初始化
func (mr *LocalGpuWorker) Init() {
	mr.gpuMiner = execute.NewGpuMiner(
		// test argvs
		"/media/yangjie/500GB/Hacash/src/github.com/hacash/x16rs/opencl",
		"Intel(R) OpenCL HD Graphics",
		false,
	)
}

// 开始采矿
func (mr *LocalGpuWorker) DoMining(blockHeight uint64, retmaxhash bool, stopmark *byte, hashstart uint32, hashend uint32, tarhashvalue []byte, blockheadmeta []byte) (byte, bool, []byte, []byte) {

	return 0, false, []byte{0, 0, 0, 0}, bytes.Repeat([]byte{255}, 32)
}
