package gpuexec

import (
	"fmt"
	"github.com/hacash/miner/device"
	"github.com/xfong/go2opencl/cl"
	"strings"
)

type GPUManage struct {
	config   *device.Config
	platform *cl.Platform
	context  *cl.Context
	program  *cl.Program
	devices  []*cl.Device
}

func NewGPUManage(cnf *device.Config) *GPUManage {
	return &GPUManage{
		config: cnf,
	}
}

func (g *GPUManage) GetDevices() []*cl.Device {
	return g.devices
}

func (g *GPUManage) Init() error {
	// start
	platforms, e := cl.GetPlatforms()
	if e != nil {
		return e
	}
	chooseplatids := 0
	platmc := g.config.GPU_PlatformNameMatch
	for i, pt := range platforms {
		fmt.Printf("  - platform %d: %s\n", i, pt.Name())
		if strings.Compare(platmc, "") != 0 && strings.Contains(pt.Name(), platmc) {
			chooseplatids = i
		}
	}
	// get platform
	g.platform = platforms[chooseplatids]
	fmt.Printf("current use platform: %s\n", g.platform.Name())
	g.devices, e = g.platform.GetDevices(cl.DeviceTypeAll)
	if len(g.devices) <= 0 || e != nil {
		fmt.Printf(fmt.Sprintf("\n--------\n-- GPU Error: %s\n--------\n", "Cannot find any GPU device!!!"))
		return e
	}
	for i, dv := range g.devices {
		fmt.Printf("  - device %d: %s, (max_work_group_size: %d)\n", i, dv.Name(), dv.MaxWorkGroupSize())
	}
	// context
	g.context, _ = cl.CreateContext(g.devices)
	// build
	g.program = BuildProgram(g.config, g.context, g.devices)

	// ok
	return nil
}
