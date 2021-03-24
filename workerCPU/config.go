package workerCPU

import (
	"github.com/hacash/core/sys"
)

type CPUWorkerConfig struct {
	Supervene     uint32 // CPU 并发挖矿
	IsReportPower bool   // 是否上报算力
}

func NewEmptyCPUWorkerConfig() *CPUWorkerConfig {
	cnf := &CPUWorkerConfig{
		Supervene: 1,
	}
	return cnf
}

//////////////////////////////////////////////////

func NewCPUWorkerConfig(cnffile *sys.Inicnf) *CPUWorkerConfig {
	cnf := NewEmptyCPUWorkerConfig()
	cnfsection := cnffile.Section("")
	// supervene
	cnf.Supervene = uint32(cnfsection.Key("supervene").MustUint(1))
	// IsReportPower
	cnf.IsReportPower = cnfsection.Key("not_report_power").MustBool(false) == false
	// ok
	return cnf
}
