package localgpu

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/mint/difficulty"
	"github.com/xfong/go2opencl/cl"

	//"github.com/hacash/x16rs"
	"sync/atomic"
)

type miningBlockReturn struct {
	stopKind       byte
	isSuccess      bool
	coinbaseMsgNum uint32
	nonceBytes     []byte
	powerHash      []byte
	blockHeadMeta  interfaces.Block
}

type GPUWorker struct {
	platform *cl.Platform
	context  *cl.Context
	program  *cl.Program
	devices  []*cl.Device // 所有设备

	deviceworkers []*GpuMinerDeviceWorkerContext

	// config
	openclPath        string
	rebuild           bool   // 强制重新编译
	platName          string // 选择的平台
	groupNum          int    // 同时执行组数量
	groupSize         int    // 组大小
	itemLoop          int    // 单次执行循环次数
	emptyFuncTest     bool   // 空函数编译测试
	useOneDeviceBuild bool   // 使用单个设备编译

	returnPowerHash bool

	stopMark *byte

	coinbaseMsgNum uint32

	successMiningMark *uint32

	successBlockCh chan miningBlockReturn
}

func NewGPUWorker(successMiningMark *uint32, successBlockCh chan miningBlockReturn, coinbaseMsgNum uint32, stopMark *byte, config *LocalGPUPowMasterConfig) *GPUWorker {
	worker := &GPUWorker{
		openclPath:        config.openclPath,
		platName:          config.platName,
		rebuild:           false,
		emptyFuncTest:     config.emptyFuncTest,
		useOneDeviceBuild: config.useOneDeviceBuild,
		groupSize:         config.groupSize,
		groupNum:          config.groupNum,
		itemLoop:          config.itemLoop,
		returnPowerHash:   false,
		successMiningMark: successMiningMark,
		successBlockCh:    successBlockCh,
		coinbaseMsgNum:    coinbaseMsgNum,
		stopMark:          stopMark,
	}
	return worker
}

func (c *GPUWorker) RunMining(newblockheadmeta interfaces.Block, startNonce uint32, endNonce uint32) bool {
	workStuff := blocks.CalculateBlockHashBaseStuff(newblockheadmeta)
	targethashdiff := difficulty.Uint32ToHash(newblockheadmeta.GetHeight(), newblockheadmeta.GetDifficulty())
	// run
	//fmt.Println( "targethashdiff:", hex.EncodeToString(targethashdiff) )
	// ========= test start =========
	//time.Sleep(time.Second)
	// ========= test end   =========
	//stopkind, issuccess, noncebytes, powerhash := x16rs.MinerNonceHashX16RS(newblockheadmeta.GetHeight(), c.returnPowerHash, c.stopMark, startNonce, endNonce, targethashdiff, workStuff)
	issuccess, noncebytes, powerhash := c.deviceworkers.DoMining(devideCtx, globalwide, groupsize, x16rsrepeat, uint32(basenoncestart))
	//fmt.Println("x16rs.MinerNonceHashX16RS finish ", issuccess,  binary.LittleEndian.Uint32(noncebytes[0:4]), startNonce, endNonce)
	if issuccess && atomic.CompareAndSwapUint32(c.successMiningMark, 0, 1) {
		// return success block
		*c.stopMark = 1 // set stop mark for all cpu worker
		//fmt.Println("start c.successBlockCh <- newblock")
		c.successBlockCh <- miningBlockReturn{
			stopkind,
			true,
			c.coinbaseMsgNum,
			noncebytes,
			nil,
			newblockheadmeta,
		}
		//fmt.Println("end ... c.successBlockCh <- newblock")
		return true
	} else if c.returnPowerHash {
		c.successBlockCh <- miningBlockReturn{
			stopkind,
			false,
			c.coinbaseMsgNum,
			noncebytes,
			powerhash,
			newblockheadmeta,
		}
		return false
	}
	return false
}

func (mr *GPUWorker) doGroupWork(ctx *GpuMinerDeviceWorkerContext, global int, local int, x16rsrepeat uint32, base_start uint32) (bool, []byte, []byte) {

	// time.Sleep(time.Millisecond * 300)

	var e error

	// 重置
	_, e = ctx.queue.EnqueueWriteBufferByte(ctx.output_nonce, true, 0, []byte{0, 0, 0, 0}, nil)
	if e != nil {
		panic(e)
	}
	// set argvs
	e = ctx.kernel.SetArgs(ctx.input_target, ctx.input_stuff, x16rsrepeat, uint32(base_start), uint32(mr.itemLoop), ctx.output_nonce, ctx.output_hash)
	if e != nil {
		panic(e)
	}
	// run
	//fmt.Println("EnqueueNDRangeKernel")
	_, e = ctx.queue.EnqueueNDRangeKernel(ctx.kernel, []int{0}, []int{global}, []int{local}, nil)
	if e != nil {
		fmt.Println("EnqueueNDRangeKernel ERROR:")
		panic(e)
	}
	//fmt.Println("EnqueueNDRangeKernel END!!!")
	//fmt.Println("ctx.queue.Finish() start")
	e = ctx.queue.Finish()
	if e != nil {
		panic(e)
	}
	//fmt.Println("ctx.queue.Finish() end")

	result_nonce := bytes.Repeat([]byte{0}, 4)
	result_hash := make([]byte, 32)
	// copy get output
	//fmt.Println("EnqueueReadBufferByte output_nonce start")
	_, e = ctx.queue.EnqueueReadBufferByte(ctx.output_nonce, true, 0, result_nonce, nil)
	if e != nil {
		panic(e)
	}
	//fmt.Println("EnqueueReadBufferByte output_nonce end")
	//fmt.Println("EnqueueReadBufferByte output_hash start")
	_, e = ctx.queue.EnqueueReadBufferByte(ctx.output_hash, true, 0, result_hash, nil)
	if e != nil {
		panic(e)
	}
	//fmt.Println("EnqueueReadBufferByte output_hash end")
	//fmt.Println("ctx.queue.Finish() start")
	e = ctx.queue.Finish()
	if e != nil {
		panic(e)
	}
	//fmt.Println("ctx.queue.Finish() end")

	// check results
	//fmt.Println("==========================", result_nonce, hex.EncodeToString(result_nonce))
	//fmt.Println("output_hash", result_hash, hex.EncodeToString(result_hash))
	//fmt.Println(result_nonce)
	nonce := binary.BigEndian.Uint32(result_nonce)
	if nonce > 0 {
		// check results
		// fmt.Println("==========================", nonce, result_nonce)
		// fmt.Println("output_hash", result_hash, hex.EncodeToString(result_hash))
		// return
		return true, result_nonce, result_hash
	}
	return false, nil, nil

}
