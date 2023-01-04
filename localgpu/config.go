package localgpu

import (
	"github.com/hacash/core/sys"
)

type LocalGPUPowMasterConfig struct {
	ReturnPowerHash   bool
	Concurrent        uint32 // Concurrent mining
	OpenclPath        string
	PlatName          string // 选择的平台
	GroupNum          int    // 同时执行组数量
	GroupSize         int    // 组大小
	ItemLoop          int    // 单次执行循环次数
	EmptyFuncTest     bool   // 空函数编译测试
	UseOneDeviceBuild bool   // 使用单个设备编译
}

func NewEmptyLocalGPUPowMasterConfig() *LocalGPUPowMasterConfig {
	cnf := &LocalGPUPowMasterConfig{
		ReturnPowerHash:   false,
		Concurrent:        1,
		OpenclPath:        "",
		PlatName:          "",
		GroupNum:          1,
		GroupSize:         1,
		ItemLoop:          10,
		EmptyFuncTest:     false,
		UseOneDeviceBuild: true,
	}
	return cnf
}

//////////////////////////////////////////////////

func NewLocalGPUPowMasterConfig(cnffile *sys.Inicnf) *LocalGPUPowMasterConfig {
	cnf := NewEmptyLocalGPUPowMasterConfig()

	return cnf

}
