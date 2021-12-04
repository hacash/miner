package minerworkerwrap

import (
	"fmt"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/x16rs/cpumining"
	"github.com/hacash/x16rs/opencl/worker"
)

// 初始化
func (g *WorkerWrap) InitStart() error {

	// 初始化设备
	if g.powdevice == nil {
		var device interfaces.PowDevice = nil
		if g.config.GPU_Enable {

			// GPU
			device = worker.NewGpuMiner(
				// test argvs
				g.config.GPU_OpenclPath,
				g.config.GPU_PlatformNameMatch,
				g.config.GPU_GroupSize,
				g.config.GPU_GroupConcurrentNum,
				g.config.GPU_ItemLoopNum,       // 单次执行循环次数
				g.config.GPU_UseOneDeviceBuild, // 强制重新编译
				g.config.GPU_ForceRebuild,      // 强制重新编译
				g.config.GPU_EmptyFuncTest,     // 空函数编译测试
			)
			fmt.Printf("startup GPU device...\n")

		} else {

			// CPU 默认
			device = cpumining.NewCPUMining(
				int(g.config.Supervene),
			)
			fmt.Printf("startup CPU device of [%d] supervene.\n",
				g.config.Supervene,
			)

		}
		g.powdevice = device
	}

	if g.powdevice == nil {
		panic("must call SetPowDevice() first.")
	}

	return g.powdevice.Init() // 初始化
}

// 设置挖矿设备端
func (g *WorkerWrap) SetPowDevice(device interfaces.PowDevice) {
	g.powdevice = device
}
