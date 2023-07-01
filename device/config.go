package device

import (
	"fmt"
	"github.com/hacash/core/sys"
	"github.com/hacash/core/sys/inicnf"
)

type Config struct {
	//PoolAddress            *net.TCPAddr
	//Rewards                fields.Address

	Concurrent uint32 // Concurrent mining
	Detail_Log bool

	GPU_Enable             bool
	GPU_OpenclPath         string
	GPU_UseMainFileContent string
	GPU_PlatformNameMatch  string
	GPU_GroupSize          int
	GPU_GroupConcurrent    int
	GPU_ItemLoopNum        int
	GPU_UseOneDeviceBuild  bool // Compile using a single device
	GPU_ForceRebuild       bool // Force recompile
	GPU_EmptyFuncTest      bool // Empty function compilation test
}

func (c *Config) IsDetailLog() bool {
	return c.Detail_Log
}

func NewEmptyMinerPoolWorkerConfig() *Config {
	cnf := &Config{
		Concurrent: 1,
	}
	return cnf
}

//////////////////////////////////////////////////

func NewConfig(cnfsection *inicnf.Section) *Config {
	cnf := NewEmptyMinerPoolWorkerConfig()
	/*
		// pool
		addr, err := net.ResolveTCPAddr("tcp", cnfsection.Key("pool").MustString(""))
		if err != nil {
			fmt.Println(err)
			panic("pool ip:port is error.")
		}
		cnf.PoolAddress = addr
		// rewards
		rwdaddr, e1 := fields.CheckReadableAddress(cnfsection.Key("rewards").MustString("1AVRuFXNFi3rdMrPH4hdqSgFrEBnWisWaS"))
		if e1 != nil {
			fmt.Println(e1)
			panic("reward address is error.")
		}
		cnf.Rewards = *rwdaddr
	*/
	// supervene
	cnf.Concurrent = uint32(cnfsection.Key("supervene").MustUint(1))
	cnf.Detail_Log = cnfsection.Key("detail_log").MustBool(false)

	//gpusection := cnffile.Section("GPU")
	gpusection := cnfsection

	cnf.GPU_Enable = gpusection.Key("gpu_enable").MustBool(false)
	cnf.GPU_OpenclPath = gpusection.Key("gpu_opencl_path").MustString("./x16rs_opencl")
	cnf.GPU_OpenclPath = sys.AbsDir(cnf.GPU_OpenclPath) // abs path for exe belong path
	fmt.Printf("[Config] load x16rs opencl dir: %s\n", cnf.GPU_OpenclPath)
	cnf.GPU_UseMainFileContent = gpusection.Key("gpu_use_main_file_content").MustString("")
	cnf.GPU_PlatformNameMatch = gpusection.Key("gpu_platform_match").MustString("")
	cnf.GPU_GroupSize = int(gpusection.Key("gpu_group_size").MustInt(32))
	cnf.GPU_GroupConcurrent = int(gpusection.Key("gpu_group_concurrent").MustInt(32))
	cnf.GPU_ItemLoopNum = int(gpusection.Key("gpu_item_loop").MustUint(0))
	cnf.GPU_UseOneDeviceBuild = gpusection.Key("gpu_use_single_device_build").MustBool(false)
	cnf.GPU_ForceRebuild = gpusection.Key("gpu_rebuild").MustBool(false)
	cnf.GPU_EmptyFuncTest = gpusection.Key("gpu_empty_func_test").MustBool(false)
	return cnf
}
