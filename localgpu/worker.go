package localgpu

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/mint/difficulty"
	"github.com/hacash/x16rs"
	"github.com/xfong/go2opencl/cl"
	"sync"

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

func NewGPUWorker(successMiningMark *uint32, successBlockCh chan miningBlockReturn, coinbaseMsgNum uint32, stopMark *byte, config *LocalGPUPowMasterConfig, platform *cl.Platform, context *cl.Context, program *cl.Program, devices []*cl.Device, deviceworkers []*GpuMinerDeviceWorkerContext) *GPUWorker {
	mr := &GPUWorker{
		platform:          platform,
		context:           context,
		program:           program,
		devices:           devices,
		deviceworkers:     deviceworkers,
		openclPath:        config.OpenclPath,
		platName:          config.PlatName,
		rebuild:           false,
		emptyFuncTest:     config.EmptyFuncTest,
		useOneDeviceBuild: config.UseOneDeviceBuild,
		groupSize:         config.GroupSize,
		groupNum:          config.GroupNum,
		itemLoop:          config.ItemLoop,
		returnPowerHash:   false,
		successMiningMark: successMiningMark,
		successBlockCh:    successBlockCh,
		coinbaseMsgNum:    coinbaseMsgNum,
		stopMark:          stopMark,
	}

	return mr
}

func (c *GPUWorker) RunMining(newblockheadmeta interfaces.Block, stopmark *byte) bool {
STARTDOMINING:
	if *stopmark == 1 {
		return false
	}
	supervene := 1
	blockheadmetasary := make([][]byte, supervene)
	oksuffnum := 0
	for {
		if *stopmark == 1 {
			return false
		}
		tarblock := newblockheadmeta
		//fmt.Println(tarblock.GetMrklRoot())
		blockheadmeatastuff := blocks.CalculateBlockHashBaseStuff(tarblock)
		blockheadmetasary[oksuffnum] = blockheadmeatastuff // block mining stuff
		oksuffnum++
		if oksuffnum == supervene {
			break // Start digging
		}
	}
	//
	//workStuff := blocks.CalculateBlockHashBaseStuff(newblockheadmeta)
	targethashdiff := difficulty.Uint32ToHash(newblockheadmeta.GetHeight(), newblockheadmeta.GetDifficulty())
	// run
	//fmt.Println( "targethashdiff:", hex.EncodeToString(targethashdiff) )
	// ========= test start =========
	//time.Sleep(time.Second)
	if *stopmark == 1 {
		return false
	}
	// ========= test end   =========DoMining
	//stopkind, issuccess, noncebytes, powerhash := x16rs.MinerNonceHashX16RS(newblockheadmeta.GetHeight(), c.returnPowerHash, c.stopMark, startNonce, endNonce, targethashdiff, workStuff)
	issuccess, _, noncebytes, powerhash := c.DoMining(newblockheadmeta.GetHeight(), stopmark, targethashdiff, blockheadmetasary)
	//fmt.Println("x16rs.MinerNonceHashX16RS finish ", issuccess,  binary.LittleEndian.Uint32(noncebytes[0:4]), startNonce, endNonce)
	if issuccess && atomic.CompareAndSwapUint32(c.successMiningMark, 0, 1) {
		// return success block
		*c.stopMark = 1 // set stop mark for all gpu worker
		//fmt.Println("start c.successBlockCh <- newblock")
		c.successBlockCh <- miningBlockReturn{
			0,
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
			0,
			false,
			c.coinbaseMsgNum,
			noncebytes,
			powerhash,
			newblockheadmeta,
		}
		return false
	}
	if *c.stopMark == 0 {
		goto STARTDOMINING
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

func (g *GPUWorker) DoMining(blockHeight uint64, stopmark *byte, tarhashvalue []byte, blockheadmeta [][]byte) (bool, byte, []byte, []byte) {

	deviceNum := len(g.devices)

	var successed bool = false
	var successMark uint32 = 0
	var successStuffIdx byte = 0
	var successNonce []byte = nil
	var successHash []byte = nil

	// 同步等待
	var syncWait = sync.WaitGroup{}
	syncWait.Add(deviceNum)

	// 设备执行
	for i := 0; i < deviceNum; i++ {
		go func(did int) {
			defer syncWait.Done()
			fmt.Println("mr.deviceworkers[i]", did, len(g.deviceworkers), g.deviceworkers)
			//devideCtx := g.deviceworkers[did]
			stuffbts := blockheadmeta[did]
			// 执行
			x16rsrepeat := uint32(x16rs.HashRepeatForBlockHeight(blockHeight))
			var basenoncestart uint64 = 1
		RUNMINING:
			// 初始化 执行环境
			//devideCtx := g.createWorkContext(did)
			devideCtx := g.deviceworkers[did]
			devideCtx.ReInit(stuffbts, tarhashvalue)
			//fmt.Println("DO RUNMINING...")
			//ttstart := time.Now()
			groupsize := g.devices[did].MaxWorkGroupSize()
			if g.groupSize > 0 {
				groupsize = int(g.groupSize)
			}
			globalwide := groupsize * g.groupNum
			overstep := globalwide * g.itemLoop // 单次挖矿 nonce 范围
			//fmt.Println(overstep, groupsize)
			success, nonce, endhash := g.doGroupWork(devideCtx, globalwide, groupsize, x16rsrepeat, uint32(basenoncestart))
			//devideCtx.Release() // 释放
			//fmt.Println("END RUNMINING:", time.Now().Unix(), time.Now().Unix() - ttstart.Unix(), success, hex.EncodeToString(nonce), hex.EncodeToString(endhash) )
			if success && atomic.CompareAndSwapUint32(&successMark, 0, 1) {
				successed = true
				*stopmark = 1
				successStuffIdx = byte(did)
				successNonce = nonce
				successHash = endhash
				// 检查是否真的成功
				blk, _, _ := blocks.ParseExcludeTransactions(stuffbts, 0)
				blk.SetNonceByte(nonce)
				nblkhx := blk.HashFresh()
				if difficulty.CheckHashDifficultySatisfy(nblkhx, tarhashvalue) == false || bytes.Compare(nblkhx, endhash) != 0 {
					fmt.Println("挖矿失败！！！！！！！！！！！！！！！！")
					fmt.Println(nblkhx.ToHex(), hex.EncodeToString(endhash))
					fmt.Println(hex.EncodeToString(stuffbts))
				}

				return // 成功挖出，结束
			}
			if *stopmark == 1 {
				//fmt.Println("ok.")
				return // 稀缺一个区块，结束
			}
			// 继续挖款
			basenoncestart += uint64(overstep)
			if basenoncestart > uint64(4294967295) {
				//if basenoncestart > uint64(529490) {
				return // 本轮挖挖矿结束
			}
			//time.Sleep(time.Second * 5)
			goto RUNMINING
		}(i)
	}

	//fmt.Println("syncWait.Wait()")
	// 等待
	syncWait.Wait()

	// 返回
	return successed, successStuffIdx, successNonce, successHash

}
