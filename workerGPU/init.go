package workerGPU

import "github.com/hacash/x16rs/opencl/execute"

// 初始化
func (g *GpuWorker) InitStart() error {
	g.gpuMiner = execute.NewGpuMiner(
		// test argvs
		g.config.GPU_OpenclPath,
		"Intel(R) OpenCL HD Graphics",
		false,
	)

	return nil
}
