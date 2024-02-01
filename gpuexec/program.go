package gpuexec

import (
	"fmt"
	"github.com/hacash/miner/device"
	cl2 "github.com/hacash/miner/gpuexec/cl"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"
)

func getBinName(device *cl2.Device) string {
	var binfilename = strings.Replace(device.Name(), " ", "_", -1) + ".bin"
	return binfilename
}

func isAllDeviceSame(devices []*cl2.Device) bool {
	var dvdnum = len(devices)
	var is_all_devide_as_same = false
	if dvdnum > 1 {
		is_all_devide_as_same = true
	}
	for i := 1; i < dvdnum; i++ {
		if devices[i].Name() != devices[0].Name() {
			is_all_devide_as_same = false
			break
		}
	}
	return is_all_devide_as_same
}

func BuildProgram(config *device.Config, platform *cl2.Platform, context *cl2.Context, devices []*cl2.Device) *cl2.Program {

	var e error
	var program *cl2.Program
	var dvdnum = len(devices)
	var is_all_devide_as_same = isAllDeviceSame(devices)

	if config.GPU_ForceRebuild {
		goto BUILD_FROM_SOURCE_FOUCE
	}

BUILD_FROM_BINARIES:

	// check binaries file
	if !config.GPU_EmptyFuncTest {
		var allbins = make([][]byte, dvdnum)
		for i := 0; i < dvdnum; i++ {
			binfilename := getBinName(devices[i])
			var binfpw = path.Join(config.GPU_OpenclPath, binfilename)
			bincon, e := os.ReadFile(binfpw)
			if e != nil {
				goto BUILD_FROM_SOURCE_WITH_SAME_CHECK
			}
			allbins[i] = bincon
		}

		fmt.Print("Create OpenCL program with binaries...")
		program, e = context.CreateProgramWithBinary(devices, allbins)
		if e != nil {
			panic(e)
		}
		e = program.BuildProgram(devices, "")
		if e != nil {
			panic(e)
		}
		fmt.Println("program from binaries create successfully.")
		return program
	}

BUILD_FROM_SOURCE_WITH_SAME_CHECK:

	if is_all_devide_as_same {
		// build one & use for all
		fmt.Printf("All %d devices are the same and only need to be compiled once...\n", dvdnum)
		is_all_devide_as_same = false // just do one
		program = buildFromSource(config, platform, context, []*cl2.Device{devices[0]})
		goto BUILD_FROM_BINARIES
	}

BUILD_FROM_SOURCE_FOUCE:

	program = buildFromSource(config, platform, context, devices)

	// 返回
	return program
}

func buildFromSource(config *device.Config, platform *cl2.Platform, context *cl2.Context, devices []*cl2.Device) *cl2.Program {

	var e error
	var program *cl2.Program

	fmt.Printf("Create OpenCL program with source for %d devices form %s, please wait...\n", len(devices), config.GPU_OpenclPath)
	buildok := false
	go func() { // 打印
		ct := 0
		for {
			time.Sleep(time.Second * 3)
			if buildok {
				break
			}
			ct += 3
			fmt.Printf("\rCompilation time: %ds ...         ", ct)
		}
	}()
	emptyFuncTest := ""
	if config.GPU_EmptyFuncTest {
		emptyFuncTest = `_empty_test` // 空函数快速编译测试
	}

	codeString := ` #include "x16rs_main` + emptyFuncTest + `.cl" `
	if !config.GPU_EmptyFuncTest && len(config.GPU_UseMainFileContent) > 0 {
		codeString = config.GPU_UseMainFileContent // use outside code content
	}
	codeString += fmt.Sprintf("\n#define updateforbuild %d", rand.Uint64()) // 避免某些平台编译缓存
	program, e = context.CreateProgramWithSource([]string{codeString})
	if e != nil {
		panic(e)
	}
	bderr := program.BuildProgram(devices, "-I "+config.GPU_OpenclPath)
	if bderr != nil {
		panic(bderr)
	}
	buildok = true // build 完成
	//fmt.Println("\nBuild complete get binaries...")

	fmt.Println("GPU miner program create complete successfully.")

	// save
	if !config.GPU_EmptyFuncTest {
		saveBinaries(config.GPU_OpenclPath, devices, program)
	}

	return program
}

func saveBinaries(dir string, dvds []*cl2.Device, program *cl2.Program) {

	binbtary, e := program.GetBinarieByDevices(dvds)
	if e != nil {
		panic(e)
	}
	// save
	os.Mkdir(dir, 0777)

	maxlen := len(dvds)
	if maxlen > len(binbtary) {
		maxlen = len(binbtary)
	}

	// save each
	for i := 0; i < maxlen; i++ {
		var fname = getBinName(dvds[i])
		var fn = path.Join(dir, fname)
		e = os.WriteFile(fn, binbtary[i], 0777)
		if e != nil {
			panic(e)
		}
	}

	//fmt.Println("WriteFile ok")
}
