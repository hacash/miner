package gpuexec

import (
	"fmt"
	"github.com/hacash/miner/device"
	cl2 "github.com/hacash/miner/gpuexec/cl"
	"strings"
)

type GPUManage struct {
	config   *device.Config
	platform *cl2.Platform
	context  *cl2.Context
	program  *cl2.Program
	devices  []*cl2.Device
}

func NewGPUManage(cnf *device.Config) *GPUManage {
	return &GPUManage{
		config: cnf,
	}
}

func (g *GPUManage) GetDevices() []*cl2.Device {
	return g.devices
}

func (g *GPUManage) Init() error {
	var e error
	// start
	platforms := cl2.GetPlatforms()

	chooseplatids := 0
	platmc := g.config.GPU_PlatformNameMatch
	for i, pt := range platforms {
		fmt.Printf("  - platform %d: %s\n", i, pt.Name())
		if strings.Compare(platmc, "") != 0 && strings.Contains(pt.Name(), platmc) {
			chooseplatids = i
		}
	}
	if len(platforms) == 0 {
		return fmt.Errorf("Cannot find any GPU platforms")
	}

	// get platform
	g.platform = platforms[chooseplatids]
	fmt.Printf("current use platform: %s\n", g.platform.Name())
	g.devices = g.platform.GetDevices(cl2.CL_DEVICE_TYPE_ALL)
	if len(g.devices) <= 0 {
		e := fmt.Sprintf("\n--------\n-- GPU Error: %s\n--------\n", "Cannot find any GPU device!!!")
		fmt.Printf(e)
		return fmt.Errorf(e)
	}
	for i, dv := range g.devices {
		fmt.Printf("  - device %d: %s, (max_work_group_size: %d)\n", i, dv.Name(), dv.MaxWorkGroupSize())
	}
	// context
	g.context, e = cl2.CreateContext(g.devices)
	if e != nil {
		fmt.Println(e.Error())
		return e
	}
	// build
	g.program = BuildProgram(g.config, g.platform, g.context, g.devices)

	// ok
	return nil
}
