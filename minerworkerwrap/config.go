package minerworkerwrap

import (
	"github.com/hacash/core/sys"
)

type WorkerWrapConfig struct {
	IsReportPower bool // 是否上报算力
	// CPU 配置
	Supervene uint32 // CPU 并发挖矿
	// GPU 配置
	GPU_Enable             bool
	GPU_OpenclPath         string
	GPU_PlatformNameMatch  string
	GPU_GroupSize          int
	GPU_GroupConcurrentNum int
	GPU_ItemLoopNum        int
	GPU_UseOneDeviceBuild  bool // 使用单个设备编译
	GPU_ForceRebuild       bool // 强制重新编译
	GPU_EmptyFuncTest      bool // 空函数编译测试
}

func NewEmptyWorkerWrapConfig() *WorkerWrapConfig {
	cnf := &WorkerWrapConfig{}
	return cnf
}

//////////////////////////////////////////////////

func NewWorkerWrapConfig(cnffile *sys.Inicnf) *WorkerWrapConfig {
	cnf := NewEmptyWorkerWrapConfig()
	cnfsection := cnffile.Section("")
	// supervene
	cnf.Supervene = uint32(cnfsection.Key("supervene").MustUint(1))
	// IsReportPower
	cnf.IsReportPower = cnfsection.Key("not_report_power").MustBool(false) == false
	// GPU
	gpusection := cnffile.Section("GPU")
	cnf.GPU_Enable = gpusection.Key("enable").MustBool(false)
	cnf.GPU_OpenclPath = gpusection.Key("opencl_path").MustString("")
	cnf.GPU_PlatformNameMatch = gpusection.Key("platform_match").MustString("")
	cnf.GPU_GroupSize = int(gpusection.Key("group_size").MustInt(0))
	cnf.GPU_GroupConcurrentNum = int(gpusection.Key("group_concurrent").MustInt(1))
	cnf.GPU_ItemLoopNum = int(gpusection.Key("group_item_loop").MustUint(10))
	cnf.GPU_UseOneDeviceBuild = gpusection.Key("use_single_device_build").MustBool(false)
	cnf.GPU_ForceRebuild = gpusection.Key("rebuild").MustBool(false)
	cnf.GPU_EmptyFuncTest = gpusection.Key("empty_func_test").MustBool(false)
	//fmt.Println("cnf.GPU_GroupConcurrentNum = = ", cnf.GPU_GroupConcurrentNum)
	// ok
	return cnf
}
