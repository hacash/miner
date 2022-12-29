package localgpu

import (
	"github.com/hacash/core/sys"
)

type LocalGPUPowMasterConfig struct {
	ReturnPowerHash   bool
	Concurrent        uint32 // Concurrent mining
	openclPath        string
	platName          string // 选择的平台
	groupNum          int    // 同时执行组数量
	groupSize         int    // 组大小
	itemLoop          int    // 单次执行循环次数
	emptyFuncTest     bool   // 空函数编译测试
	useOneDeviceBuild bool   // 使用单个设备编译
}

func NewEmptyLocalGPUPowMasterConfig() *LocalGPUPowMasterConfig {
	cnf := &LocalGPUPowMasterConfig{
		ReturnPowerHash: false,
		Concurrent:      1,
	}
	return cnf
}

//////////////////////////////////////////////////

func NewLocalGPUPowMasterConfig(cnffile *sys.Inicnf) *LocalGPUPowMasterConfig {
	cnf := NewEmptyLocalGPUPowMasterConfig()

	return cnf

}
