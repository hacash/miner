package localgpu

import (
	"fmt"
	"github.com/xfong/go2opencl/cl"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"
)

type GpuMinerDeviceWorkerContext struct {
	context      *cl.Context
	device       *cl.Device
	kernel       *cl.Kernel
	queue        *cl.CommandQueue
	input_stuff  *cl.MemObject
	input_target *cl.MemObject
	output_nonce *cl.MemObject
	output_hash  *cl.MemObject
}

func (w *GpuMinerDeviceWorkerContext) Retain() {
	w.kernel.Retain()
	w.queue.Retain()
	w.input_stuff.Retain()
	w.input_target.Retain()
	w.output_nonce.Retain()
	w.output_hash.Retain()
}

func (w *GpuMinerDeviceWorkerContext) ReInit(stuff_buf []byte, target_buf []byte) {
	// set
	w.queue.EnqueueWriteBufferByte(w.input_stuff, true, 0, stuff_buf, nil)
	w.queue.EnqueueWriteBufferByte(w.input_target, true, 0, target_buf, nil)
}

// chua
func (mr *LocalGPUPowMaster) createWorkContext(devidx int) *GpuMinerDeviceWorkerContext {

	// 运行创建执行单元
	//input_target_buf := make([]byte, 32)
	//copy(input_target_buf, work.target[:])
	//input_stuff_buf := make([]byte, 89)
	//copy(input_stuff_buf, work.blkstuff[:])
	// |cl.MemCopyHostPtr
	input_target, _ := mr.context.CreateEmptyBuffer(cl.MemReadOnly, 32)
	input_stuff, _ := mr.context.CreateEmptyBuffer(cl.MemReadOnly, 89)
	//defer input_target.Release()
	//defer input_stuff.Release()

	// 参数
	// |cl.MemAllocHostPtr
	output_nonce, _ := mr.context.CreateEmptyBuffer(cl.MemReadWrite|cl.MemAllocHostPtr, 4)
	output_hash, _ := mr.context.CreateEmptyBuffer(cl.MemReadWrite|cl.MemAllocHostPtr, 32)
	//defer output_nonce.Release()
	//defer output_hash.Release()

	kernel, ke1 := mr.program.CreateKernel("miner_do_hash_x16rs_v2")
	if ke1 != nil {
		panic(ke1)
	}
	//fmt.Println("mr.program.CreateKernel SUCCESS")
	//defer kernel.Release()

	device := mr.devices[devidx]
	queue, qe1 := mr.context.CreateCommandQueue(device, 0)
	if qe1 != nil {
		panic(qe1)
	}
	//defer queue.Release()

	ctx := &GpuMinerDeviceWorkerContext{
		mr.context,
		device,
		kernel,
		queue,
		input_stuff,
		input_target,
		output_nonce,
		output_hash,
	}

	// 复用
	ctx.Retain()

	// 返回
	return ctx

}
func (mr *LocalGPUPowMaster) buildOrLoadProgram() *cl.Program {

	var program *cl.Program

	binfilestuff := mr.platform.Name() // + "_" + mr.devices[0].Name()
	binfilename := strings.Replace(binfilestuff, " ", "_", -1)
	binfilepath := mr.config.OpenclPath + "/" + binfilename + ".objcache"
	binstat, staterr := os.Stat(binfilepath)
	if staterr != nil {
		fmt.Print("Create opencl program with source: " + mr.config.OpenclPath + ", Please wait...")
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
		if mr.config.EmptyFuncTest {
			emptyFuncTest = `_empty_test` // 空函数快速编译测试
		}
		codeString := ` #include "x16rs_main` + emptyFuncTest + `.cl" `
		codeString += fmt.Sprintf("\n#define updateforbuild %d", rand.Uint64()) // 避免某些平台编译缓存
		program, _ = mr.context.CreateProgramWithSource([]string{codeString})
		bderr := program.BuildProgram(mr.devices, "-I "+mr.config.OpenclPath)
		if bderr != nil {
			panic(bderr)
		}
		buildok = true // build 完成
		fmt.Println("\nBuild complete get binaries...")
		/*
		//fmt.Println("program.GetBinarySizes_2()")
		//size := len(mr.devices)
		sizes, _ := program.GetBinarySizes()
		//fmt.Println(sizes)
		//fmt.Println("GetBinarySizes", sizes[0])
		//fmt.Println("program.GetBinaries()")
		bins, _ := program.GetBinaries()
		binsary := make([][]uint8, len(bins))
		for i := 0; i < len(bins); i++ {
			var bisbts = make([]uint8, sizes[i])
			*&bisbts[0] = *bins[i]
			binsary[i] = bisbts
		}
		//fmt.Println("bins[0].size", len(bins[0]))
		f, e := os.OpenFile(binfilepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
		if e != nil {
			panic(e)
		}
		//fmt.Println("f.Write(wbin) "+binfilepath, sizes[0])
		var berr error
		_, berr = f.Write(binsary[0])
		if berr != nil {
			panic(berr)
		}
		berr = f.Close()
		if berr != nil {
			panic(berr)
		}
		*/

	} else {
		fmt.Printf("Load binary program file from \"%s\"\n", binfilepath)
		file, _ := os.OpenFile(binfilepath, os.O_RDONLY, 0777)
		bin := make([]byte, 0)
		//fmt.Println("file.Read(bin) size", binstat.Size())
		var berr error
		bin, berr = ioutil.ReadAll(file)
		if berr != nil {
			panic(berr)
		}
		if int64(len(bin)) != binstat.Size() {
			panic("int64(len(bin)) != binstat.Size()")
		}
		berr = file.Close()
		if berr != nil {
			panic(berr)
		}
		//fmt.Println(bin)
		// 仅仅支持同一个平台的同一种设备
		bins := make([]*uint8, len(mr.devices))
		sizes := make([]int, len(mr.devices))
		for k, _ := range mr.devices {
			bins[k] = &bin[0]
			sizes[k] = int(binstat.Size())
		}
		fmt.Println("Create program with binary...")
		program, berr = mr.context.CreateProgramWithBinary(mr.devices, sizes, bins)
		if berr != nil {
			panic(berr)
		}
		err := program.BuildProgram(mr.devices, "")
		if err != nil {
			panic(berr)
		}
		//fmt.Println("context.CreateProgramWithBinary")
	}
	fmt.Println("GPU miner program create complete successfully.")

	// 返回
	return program
}
