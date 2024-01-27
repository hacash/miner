package gpuexec

import (
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/miner/device"
	cl2 "github.com/hacash/miner/gpuexec/cl"
	itfcs "github.com/hacash/miner/interfaces"
	"time"
)

type GPUExecute struct {
	config     *device.Config
	nonce_span uint32
	allotr     chan itfcs.PoWExecute
	// GPU
	gpumng     *GPUManage
	gpucontext *ExecuteContext
}

func NewGPUExecute(cnf *device.Config) *GPUExecute {
	var span = cnf.GPU_GroupSize * cnf.GPU_GroupConcurrent * 2
	return &GPUExecute{
		config:     cnf,
		nonce_span: uint32(span),
	}
}

func (c *GPUExecute) Config() itfcs.PoWConfig {
	return c.config
}

func (c *GPUExecute) CreateContext(gpumng *GPUManage, dvc *cl2.Device) {
	var group_quanity = c.config.GPU_GroupConcurrent
	c.gpucontext = CreateExecuteContext(gpumng.program,
		gpumng.context, dvc, group_quanity)
}

func (c *GPUExecute) Allocate() chan itfcs.PoWExecute {
	return c.allotr
}

func (c *GPUExecute) StartAllocate() {
	c.gpumng = NewGPUManage(c.config)
	c.gpumng.Init()
	var dvs = c.gpumng.GetDevices()
	if len(dvs) <= 0 {
		fmt.Println("Cannot find any GPU device !")
		return
	}
	c.allotr = make(chan itfcs.PoWExecute, len(dvs))
	go func() {
		var count = 0
		for {
			exec := NewGPUExecute(c.config)
			exec.CreateContext(c.gpumng, dvs[count])
			c.allotr <- exec
			count++
			if count >= len(dvs) {
				// Concurrent be get max
				close(c.allotr)
				break
			}
		}
		//fmt.Printf("GPU Miner Concurrent %d \n", count)
	}()
}

// second
func (c *GPUExecute) ReportSpanTime(sec float64) {
	if c.config.GPU_ItemLoopNum > 0 {
		// force use config
		return
	}
	// setting span
	var span_sec = c.config.DeviceSpanTime // 10.0
	var cnfmin = float64(c.config.GPU_GroupSize * c.config.GPU_GroupConcurrent)
	var old = float64(c.nonce_span)
	var nrpt = old / (sec / span_sec) // base 10 second
	var min = old / 4
	var max = old * 4
	if nrpt < min {
		nrpt = min
	}
	if nrpt < cnfmin {
		nrpt = cnfmin
	}
	if nrpt > max {
		nrpt = max
	}
	if nrpt > 4294967294 {
		nrpt = 4294967294
	}
	// auto change
	c.nonce_span = uint32(nrpt)
}

func (c *GPUExecute) GetNonceSpan() uint32 {
	if c.config.GPU_ItemLoopNum > 0 {
		return uint32(c.config.GPU_ItemLoopNum * c.config.GPU_GroupSize * c.config.GPU_GroupConcurrent)
	}
	return c.nonce_span
}

func (c *GPUExecute) Init() error {
	return nil
}

func (c *GPUExecute) DoMining(stopmark *byte, successmark *byte, input interfaces.Block, nonce_offset uint32) (*itfcs.PoWResultData, error) {
	var block_height = input.GetHeight()
	var result = &itfcs.PoWResultData{
		PoWResultShortData: itfcs.PoWResultShortData{
			FindSuccess:   fields.CreateBool(false),
			BlockHeight:   fields.BlockHeight(block_height),
			BlockNonce:    fields.VarUint4(0),
			CoinbaseNonce: nil,
		},
	}

	// do mining
	//var nonce_use = nonce_offset
	//var nonce_max = nonce_offset + c.nonce_span
	var item_loop = float64(c.GetNonceSpan()) /
		float64(c.config.GPU_GroupSize) /
		float64(c.config.GPU_GroupConcurrent)
	if c.config.Detail_Log {
		fmt.Printf("-%d ", int64(item_loop))
	}

	var mining_end_ch = make(chan int, 3)
	var err error = nil
	var mining_span_end = false

	// gpu do
	go func() {
		var e error = nil
		r1, r2, e := c.gpucontext.DoMining(c.config, input, nonce_offset, uint32(item_loop))
		if e != nil {
			err = e
			result = nil // nothing
			mining_end_ch <- -1
			return
		}
		if result != nil {
			// ok
			mining_span_end = true
			result.ResultHash, result.BlockNonce = r1, r2
			mining_end_ch <- 1
		}
	}()

	// listen successmark
	go func() {
		for {
			time.Sleep(time.Millisecond * 250)
			if *stopmark == 1 && mining_span_end {
				return // end
			}
			if *successmark != 1 {
				continue
			}
			result = nil // nothing
			mining_end_ch <- 0
			return
		}
	}()

	// wait finish
	<-mining_end_ch
	close(mining_end_ch)

	// end
	return result, err
}
