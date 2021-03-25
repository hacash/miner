package workerGPU

import (
	"github.com/hacash/core/sys"
)

type GpuWorkerConfig struct {
	IsReportPower bool // 是否上报算力
	// GPU 配置
	GPU_OpenclPath string
}

func NewEmptyGpuWorkerConfig() *GpuWorkerConfig {
	cnf := &GpuWorkerConfig{}
	return cnf
}

//////////////////////////////////////////////////

func NewGpuWorkerConfig(cnffile *sys.Inicnf) *GpuWorkerConfig {
	cnf := NewEmptyGpuWorkerConfig()
	// GPU
	gpusection := cnffile.Section("GPU")
	cnf.GPU_OpenclPath = gpusection.Key("opencl_path").MustString("")
	// ok
	return cnf
}
