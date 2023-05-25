package gpuexec

import (
	"fmt"
	"github.com/hacash/miner/device"
	"github.com/xfong/go2opencl/cl"
	"math/rand"
	"time"
)

func BuildProgram(config *device.Config, context *cl.Context, devices []*cl.Device) *cl.Program {

	var program *cl.Program

	fmt.Print("Create opencl program with source: " + config.GPU_OpenclPath + ", Please wait...")
	buildok := false
	go func() { // 打印
		for {
			time.Sleep(time.Second * 3)
			if buildok {
				break
			}
			fmt.Print(".")
		}
	}()
	emptyFuncTest := ""
	if config.GPU_EmptyFuncTest {
		emptyFuncTest = `_empty_test` // 空函数快速编译测试
	}

	codeString := ` #include "x16rs_main` + emptyFuncTest + `.cl" `
	codeString += fmt.Sprintf("\n#define updateforbuild %d", rand.Uint64()) // 避免某些平台编译缓存
	program, _ = context.CreateProgramWithSource([]string{codeString})
	bderr := program.BuildProgram(devices, "-I "+config.GPU_OpenclPath)
	if bderr != nil {
		panic(bderr)
	}
	buildok = true // build 完成
	//fmt.Println("\nBuild complete get binaries...")

	fmt.Println("GPU miner program create complete successfully.")

	// 返回
	return program
}
