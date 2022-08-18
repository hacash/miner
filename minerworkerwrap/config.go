package minerworkerwrap

import (
	"github.com/hacash/core/sys"
)

type WorkerWrapConfig struct {
	IsReportHashrate bool // Whether to report the calculation force
	// CPU configuration
	Supervene uint32 // CPU concurrent mining
	// GPU configuration
	GPU_Enable             bool
	GPU_OpenclPath         string
	GPU_PlatformNameMatch  string
	GPU_GroupSize          int
	GPU_GroupConcurrentNum int
	GPU_ItemLoopNum        int
	GPU_UseOneDeviceBuild  bool // Compile using a single device
	GPU_ForceRebuild       bool // Force recompile
	GPU_EmptyFuncTest      bool // Empty function compilation test
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
	// IsReportHashrate
	cnf.IsReportHashrate = cnfsection.Key("not_report_hashrate").MustBool(false) == false
	// GPU
	gpusection := cnffile.Section("GPU")
	cnf.GPU_Enable = gpusection.Key("enable").MustBool(false)
	cnf.GPU_OpenclPath = gpusection.Key("opencl_path").MustString("")
	cnf.GPU_PlatformNameMatch = gpusection.Key("platform_match").MustString("")
	cnf.GPU_GroupSize = int(gpusection.Key("group_size").MustInt(1))
	cnf.GPU_GroupConcurrentNum = int(gpusection.Key("group_concurrent").MustInt(1))
	cnf.GPU_ItemLoopNum = int(gpusection.Key("group_item_loop").MustUint(10))
	cnf.GPU_UseOneDeviceBuild = gpusection.Key("use_single_device_build").MustBool(false)
	cnf.GPU_ForceRebuild = gpusection.Key("rebuild").MustBool(false)
	cnf.GPU_EmptyFuncTest = gpusection.Key("empty_func_test").MustBool(false)
	//fmt.Println("cnf.GPU_GroupConcurrentNum = = ", cnf.GPU_GroupConcurrentNum)
	// ok
	return cnf
}
