package device

import (
	"bytes"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	itfcs "github.com/hacash/miner/interfaces"
)

type CPUExecute struct {
	config     *Config
	nonce_span uint32
	allotr     chan itfcs.PoWExecute
}

func NewCPUExecute(cnf *Config) *CPUExecute {
	return &CPUExecute{
		config:     cnf,
		nonce_span: 400,
	}
}

func (c *CPUExecute) Allocate() chan itfcs.PoWExecute {
	return c.allotr
}

func (c *CPUExecute) Config() itfcs.PoWConfig {
	return c.config
}

func (c *CPUExecute) StartAllocate() {
	c.allotr = make(chan itfcs.PoWExecute)
	go func() {
		var count uint32 = 0
		for {
			c.allotr <- NewCPUExecute(c.config)
			count++
			if count >= c.config.Concurrent {
				// Concurrent be get max
				close(c.allotr)
				break
			}
		}
		// fmt.Printf("CPU Concurrent %d ", count)
	}()
}

// second
func (c *CPUExecute) ReportSpanTime(sec float64) {
	var span_sec = c.config.GPU_SpanTime // 5.0
	var old = float64(c.nonce_span)
	var nrpt = old / (sec / span_sec) // base 10 second
	var min = old / 4
	var max = old * 4
	if nrpt < min {
		nrpt = min
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

func (c *CPUExecute) GetNonceSpan() uint32 {
	return c.nonce_span
}

func (c *CPUExecute) Init() error {
	c.nonce_span = 400
	return nil
}

func (c *CPUExecute) DoMining(stopmark *byte, successmark *byte, input interfaces.Block, nonce_offset uint32) (*itfcs.PoWResultData, error) {
	var block_height = input.GetHeight()
	var result = itfcs.PoWResultData{
		PoWResultShortData: itfcs.PoWResultShortData{
			FindSuccess:   fields.CreateBool(false),
			BlockHeight:   fields.BlockHeight(block_height),
			BlockNonce:    fields.VarUint4(0),
			CoinbaseNonce: nil,
		},
	}

	// do mining
	var nonce_use = nonce_offset
	var nonce_max = nonce_offset + c.GetNonceSpan()

	var ret_hash fields.Hash = nil
	var ret_nonce uint32 = 0

	for {
		input.SetNonce(nonce_use)
		var reshx = input.HashFresh()
		if ret_hash == nil || bytes.Compare(ret_hash, reshx) == 1 {
			ret_hash = reshx // find a small hash
			ret_nonce = nonce_use
		}
		// next
		nonce_use += 1
		if nonce_use >= nonce_max {
			// finish
			break
		}
		if *stopmark == 1 {
			break
		}
	}

	// set data
	result.ResultHash = ret_hash
	result.BlockNonce = fields.VarUint4(ret_nonce)

	// end
	return &result, nil
}
