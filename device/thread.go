package device

import (
	"bytes"
	"fmt"
	"github.com/hacash/core/fields"
	itfcs "github.com/hacash/miner/interfaces"
	"github.com/hacash/mint/difficulty"
	"math/big"
	"time"
)

type PoWThreadMng struct {
	config   itfcs.PoWConfig
	executer itfcs.PoWExecute
}

var hxrate_show_count int64 = 0
var hxrate_show_ttvalue *big.Int = nil

func NewPoWThreadMng(exec itfcs.PoWExecute) *PoWThreadMng {
	return &PoWThreadMng{
		config:   exec.Config(),
		executer: exec,
	}
}

func (c *PoWThreadMng) Config() itfcs.PoWConfig {
	return c.config
}

func (c *PoWThreadMng) Init() error {
	c.executer.Init()
	return nil
}

func (c *PoWThreadMng) StopMining() {

}

func (c *PoWThreadMng) DoMining(stopmark *byte, target_hash fields.Hash, input itfcs.PoWStuffBriefData, resCh chan *itfcs.PoWResultData) error {

	var reserr error = nil
	var result *itfcs.PoWResultData = nil

	// nonce_span
	var nonce_start uint32 = 0

	var res_hash_diff *fields.Hash = nil

	for {
		if *stopmark == 1 {
			break
		}
		// do mining
		var nonce_span = c.executer.GetNonceSpan()
		start_time := time.Now()
		restep, e := c.executer.DoMining(stopmark, input.BlockHeadMeta, nonce_start)
		if e != nil {
			reserr = e
			break
		}
		if restep == nil {
			break
		}
		// check success
		if bytes.Compare(target_hash, restep.ResultHash) == 1 {
			// success find a block !
			restep.FindSuccess = fields.CreateBool(true)
			result = restep
			fmt.Printf(" \n--------\n[⬤◆◆] Successfully mined a block <%d, %s>\n--------\n", input.BlockHeadMeta.GetHeight(), restep.ResultHash.ToHex())
			break
		}
		var exec_time = time.Since(start_time).Seconds()

		// fmt.Println("exec_time----", exec_time, "----nonce_span----", nonce_span, "----result_hash----", restep.ResultHash.ToHex()[0:16])
		if c.config.IsDetailLog() {
			hxrate_show_count++
			var curhrs = difficulty.ConvertHashToRate(uint64(restep.BlockHeight), restep.ResultHash, int64(exec_time))
			if hxrate_show_ttvalue == nil {
				hxrate_show_ttvalue = curhrs
			} else {
				hxrate_show_ttvalue = hxrate_show_ttvalue.Add(hxrate_show_ttvalue, curhrs)
			}
			var curhrshow = difficulty.ConvertPowPowerToShowFormat(curhrs)
			var tthrsshow = difficulty.ConvertPowPowerToShowFormat(big.NewInt(0).Div(hxrate_show_ttvalue, big.NewInt(hxrate_show_count)))
			fmt.Printf("%d,%d %.2fs %s... %s, %s\n",
				nonce_span, nonce_start, exec_time, restep.ResultHash.ToHex()[0:20],
				curhrshow, tthrsshow)
		} else {
			fmt.Printf(".")
		}

		c.executer.ReportSpanTime(exec_time) // report exec time
		// diff hash
		if res_hash_diff == nil || bytes.Compare(*res_hash_diff, restep.ResultHash) == 1 {
			// find small hash
			res_hash_diff = &restep.ResultHash
			result = restep
		}
		// check end
		if uint64(nonce_start)+uint64(nonce_span) > 4294967294 {
			//if(uint64(nonce_start) + uint64(nonce_span) > 1000000){
			// end this uint32 nonce space
			break
		}
		// next loop
		nonce_start += nonce_span
	}

	// set data
	if result != nil {
		result.CoinbaseNonce = input.CoinbaseNonce
	}

	//fmt.Println("up:::::", result.BlockNonce,
	//	result.CoinbaseNonce.ToHex(),
	//	input.BlockHeadMeta.GetHeight(),
	//	input.BlockHeadMeta.GetMrklRoot().ToHex())

	// end
	//fmt.Println("-----------------", result)

	resCh <- result
	return reserr
}
