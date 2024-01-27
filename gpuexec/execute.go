package gpuexec

import (
	"bytes"
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/miner/device"
	cl2 "github.com/hacash/miner/gpuexec/cl"
	"github.com/hacash/mint/difficulty"
	"github.com/hacash/x16rs"
)

type ExecuteContext struct {
	context *cl2.Context
	device  *cl2.Device
	kernel  *cl2.Kernel
	queue   *cl2.CommandQueue

	input_stuff  *cl2.MemObject
	input_target *cl2.MemObject

	interim_nonce_datas *cl2.MemObject
	interim_hash_datas  *cl2.MemObject

	group_quantity int
	current_height uint64
	x16rs_repeat   int
}

func (e *ExecuteContext) Retain() {
	e.kernel.Retain()
	e.queue.Retain()
	e.input_target.Retain()
	e.input_stuff.Retain()
	e.interim_nonce_datas.Retain()
	e.interim_hash_datas.Retain()
}

func (e *ExecuteContext) DoMining(cnf *device.Config, input interfaces.Block,
	nonce_offset uint32, item_loop uint32) ([]byte, fields.VarUint4, error) {
	// do mining
	var err error = nil
	var tar_hei = input.GetHeight()
	if e.current_height != tar_hei {
		e.current_height = tar_hei
		e.x16rs_repeat = x16rs.HashRepeatForBlockHeight(tar_hei)
		target_buf := difficulty.DifficultyUint32ToHash(input.GetDifficulty())
		if err != nil {
			panic(err)
		}
		_, err = e.queue.EnqueueWriteBufferByte(e.input_target, true, 0, target_buf)
		if err != nil {
			panic(err)
		}
	}
	stuff_buf, err := input.SerializeExcludeTransactions()
	_, err = e.queue.EnqueueWriteBufferByte(e.input_stuff, true, 0, stuff_buf)
	if err != nil {
		panic(err)
	}
	// query
	//_, err := e.queue.EnqueueWriteBufferByte(e.output_nonce, true, 0, []byte{0, 0, 0, 0}, nil)
	//if err != nil {
	//	panic(err)
	//}
	err = e.queue.Finish()
	// call args
	err = e.kernel.SetArgs(e.input_target, e.input_stuff, uint32(e.x16rs_repeat), nonce_offset, item_loop,
		e.interim_nonce_datas, e.interim_hash_datas)
	if err != nil {
		panic(err)
	}
	// do call gpu
	var local = cnf.GPU_GroupSize
	var global = local * cnf.GPU_GroupConcurrent
	//fmt.Println("EnqueueNDRangeKernel do: ", global, local, uint32(e.x16rs_repeat), nonce_offset, item_loop, cnf.GPU_GroupConcurrent, cnf.GPU_GroupSize)
	_, err = e.queue.EnqueueNDRangeKernel(e.kernel, []int{0}, []int{global}, []int{local}, nil)
	if err != nil {
		fmt.Println("EnqueueNDRangeKernel ERROR:")
		panic(err)
	}
	//fmt.Println("EnqueueNDRangeKernel end")
	// exec
	err = e.queue.Finish()
	//fmt.Println("e.queue.Finish end")
	// results
	result_nonce := bytes.Repeat([]byte{0}, 4*e.group_quantity)
	result_hash := make([]byte, 32*e.group_quantity)
	// copy get output
	//fmt.Println("EnqueueReadBufferByte output_nonce start")
	_, err = e.queue.EnqueueReadBufferByte(e.interim_nonce_datas, true, 0, result_nonce)
	if err != nil {
		panic(err)
	}
	//fmt.Println("EnqueueReadBufferByte output_nonce end")
	//fmt.Println("EnqueueReadBufferByte output_hash start")
	_, err = e.queue.EnqueueReadBufferByte(e.interim_hash_datas, true, 0, result_hash)
	if err != nil {
		panic(err)
	}
	//fmt.Println("EnqueueReadBufferByte output_hash end")
	//fmt.Println("ctx.queue.Finish() start")
	err = e.queue.Finish()
	if err != nil {
		fmt.Println("queue.Finish 2 ERROR:")
		panic(err)
	}
	// ok end
	most_hash, most_nonce := getMostHashAndNonce(result_hash, result_nonce)
	return most_hash, most_nonce, nil
}

func getMostHashAndNonce(hash_datas []byte, nonce_datas []byte) (fields.Hash, fields.VarUint4) {
	var res_hash fields.Hash = nil
	var res_nonce []byte = nil
	for i := 0; i < len(hash_datas)/32; i++ {
		var st1 = i * 32
		var st2 = i * 4
		var hx = hash_datas[st1 : st1+32]
		//fmt.Println(i, hex.EncodeToString(hx))
		if i == 0 || bytes.Compare(res_hash, hx) == 1 {
			res_hash = hx
			res_nonce = nonce_datas[st2 : st2+4]
		}
	}

	//fmt.Println("------------", hex.EncodeToString(res_hash))
	result_nonce_num := fields.VarUint4(0)
	result_nonce_num.Parse(res_nonce, 0)
	return res_hash, result_nonce_num
}

// chua
func CreateExecuteContext(
	program *cl2.Program,
	context *cl2.Context,
	device *cl2.Device,
	group_quantity int) *ExecuteContext {

	// 运行创建执行单元
	//input_target_buf := make([]byte, 32)
	//copy(input_target_buf, work.target[:])
	//input_stuff_buf := make([]byte, 89)
	//copy(input_stuff_buf, work.blkstuff[:])
	// |cl.MemCopyHostPtr
	input_target, _ := context.CreateEmptyBuffer(cl2.CL_MEM_READ_ONLY, 32)
	input_stuff, _ := context.CreateEmptyBuffer(cl2.CL_MEM_READ_ONLY, 89)
	//defer input_target.Release()
	//defer input_stuff.Release()

	// |cl.MemAllocHostPtr
	var tys = cl2.CL_MEM_READ_WRITE | cl2.CL_MEM_ALLOC_HOST_PTR
	interim_nonce_datas, _ := context.CreateEmptyBuffer(tys, 4*group_quantity)
	interim_hash_datas, _ := context.CreateEmptyBuffer(tys, 32*group_quantity)

	kernel, ke1 := program.CreateKernel("miner_do_hash_x16rs_v2")
	if ke1 != nil {
		panic(ke1)
	}
	//fmt.Println("mr.program.CreateKernel SUCCESS")
	//defer kernel.Release()

	queue, qe1 := context.CreateCommandQueue(device, 0)
	if qe1 != nil {
		panic(qe1)
	}
	//defer queue.Release()

	ctx := &ExecuteContext{
		context,
		device,
		kernel,
		queue,
		input_stuff,
		input_target,
		interim_nonce_datas,
		interim_hash_datas,
		group_quantity,
		0,
		0,
	}

	// 返回
	return ctx

}
