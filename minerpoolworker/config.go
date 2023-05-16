package minerpoolworker

import (
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/sys"
	"net"
)

type MinerWorkerConfig struct {
	PoolAddress            *net.TCPAddr
	Concurrent             uint32 // Concurrent mining
	Rewards                fields.Address
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

func NewEmptyMinerPoolWorkerConfig() *MinerWorkerConfig {
	cnf := &MinerWorkerConfig{
		Concurrent: 1,
	}
	return cnf
}

//////////////////////////////////////////////////

func NewMinerWorkerConfig(cnffile *sys.Inicnf) *MinerWorkerConfig {
	cnf := NewEmptyMinerPoolWorkerConfig()
	cnfsection := cnffile.Section("")
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
	// supervene
	cnf.Concurrent = uint32(cnfsection.Key("supervene").MustUint(1))
	gpusection := cnffile.Section("GPU")
	cnf.GPU_Enable = gpusection.Key("enable").MustBool(false)
	cnf.GPU_OpenclPath = gpusection.Key("opencl_path").MustString("")
	cnf.GPU_PlatformNameMatch = gpusection.Key("platform_match").MustString("")
	cnf.GPU_GroupSize = int(gpusection.Key("group_size").MustInt(128))
	cnf.GPU_GroupConcurrentNum = int(gpusection.Key("group_concurrent").MustInt(100))
	cnf.GPU_ItemLoopNum = int(gpusection.Key("group_item_loop").MustUint(10))
	cnf.GPU_UseOneDeviceBuild = gpusection.Key("use_single_device_build").MustBool(false)
	cnf.GPU_ForceRebuild = gpusection.Key("rebuild").MustBool(false)
	cnf.GPU_EmptyFuncTest = gpusection.Key("empty_func_test").MustBool(false)
	return cnf
}
