package minerworkerwrap

import (
	"fmt"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/x16rs/cpumining"
	"github.com/hacash/x16rs/opencl/worker"
)

// initialization
func (g *WorkerWrap) InitStart() error {

	// Initialize device
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
				g.config.GPU_ItemLoopNum,       // Number of single execution cycles
				g.config.GPU_UseOneDeviceBuild, // Force recompile
				g.config.GPU_ForceRebuild,      // Force recompile
				g.config.GPU_EmptyFuncTest,     // Empty function compilation test
			)
			fmt.Printf("startup GPU device...\n")

		} else {

			// CPU default
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

	return g.powdevice.Init() // initialization
}

// Set mining equipment end
func (g *WorkerWrap) SetPowDevice(device interfaces.PowDevice) {
	g.powdevice = device
}
