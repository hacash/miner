package localgpu

import (
	"fmt"
	"github.com/hacash/core/interfaces"
	wr "github.com/hacash/x16rs/opencl/worker"
	"github.com/xfong/go2opencl/cl"
	"os"
	"strings"
	"sync"
)

type LocalGPUPowMaster struct {
	platform *cl.Platform
	context  *cl.Context
	program  *cl.Program
	devices  []*cl.Device // 所有设备

	deviceworkers []*GpuMinerDeviceWorkerContext

	config *LocalGPUPowMasterConfig

	coinbaseMsgNum uint32

	//currentWorkers mapset.Set
	stopMarks sync.Map

	stepLock sync.RWMutex

	miningStuffCh chan interfaces.PowWorkerMiningStuffItem
	resultCh      chan interfaces.PowWorkerMiningStuffItem
}

func NewLocalGPUPowMaster(cnf *LocalGPUPowMasterConfig) *LocalGPUPowMaster {

	miner := &LocalGPUPowMaster{
		config: cnf,
	}
	var e error = nil
	// opencl file prepare
	if strings.Compare(miner.config.OpenclPath, "") == 0 {
		tardir := wr.GetCurrentDirectory() + "/opencl/"
		if _, err := os.Stat(tardir); err != nil {
			fmt.Println("Create opencl dir and render files...")
			//files := wr.GetRenderCreateAllOpenclFiles() // 输出所有文件
			//err := wr.WriteClFiles(tardir, files)
			if err != nil {
				fmt.Println(e)
				os.Exit(0) // 致命错误
			}
			fmt.Println("all file ok.")
		} else {
			fmt.Println("Opencl dir already here.")
		}
		miner.config.OpenclPath = tardir
	}

	// start
	platforms, e := cl.GetPlatforms()

	chooseplatids := 0
	for i, pt := range platforms {
		fmt.Printf("  - platform %d: %s\n", i, pt.Name())
		if strings.Compare(miner.config.PlatName, "") != 0 && strings.Contains(pt.Name(), miner.config.PlatName) {
			chooseplatids = i
		}
	}

	miner.platform = platforms[chooseplatids]
	fmt.Printf("current use platform: %s\n", miner.platform.Name())

	devices, _ := miner.platform.GetDevices(cl.DeviceTypeAll)

	for i, dv := range devices {
		fmt.Printf("  - device %d: %s, (max_work_group_size: %d)\n", i, dv.Name(), dv.MaxWorkGroupSize())
	}

	// 是否单设备编译
	if miner.config.UseOneDeviceBuild {
		fmt.Println("Only use single device to build and run.")
		miner.devices = []*cl.Device{devices[0]} // 使用单台设备
	} else {
		miner.devices = devices
	}

	miner.context, _ = cl.CreateContext(miner.devices)

	// 编译源码
	miner.program = miner.buildOrLoadProgram()

	// 初始化执行环境
	devlen := len(miner.devices)
	miner.deviceworkers = make([]*GpuMinerDeviceWorkerContext, devlen)
	for i := 0; i < devlen; i++ {
		miner.deviceworkers[i] = miner.createWorkContext(i)
	}
	return miner
}

func (l *LocalGPUPowMaster) SetCoinbaseMsgNum(coinbaseMsgNum uint32) {
	l.stepLock.Lock()
	l.coinbaseMsgNum = coinbaseMsgNum
	l.stepLock.Unlock()
}
