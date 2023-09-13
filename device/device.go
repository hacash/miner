package device

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/fields"
	itfcs "github.com/hacash/miner/interfaces"
	"github.com/hacash/mint/difficulty"
	"math/big"
	"strings"
	"sync"
	"time"
)

var hxrate_show_count int64 = 0
var hxrate_show_ttvalue *big.Int = nil

type PoWDeviceMng struct {
	config  itfcs.PoWConfig
	alloter itfcs.PoWExecute
	threads []itfcs.PoWThread
}

func NewPoWDeviceMng(alloter itfcs.PoWExecute) *PoWDeviceMng {
	return &PoWDeviceMng{
		config:  alloter.Config(),
		alloter: alloter,
		threads: make([]itfcs.PoWThread, 0),
	}
}

func (c *PoWDeviceMng) Config() itfcs.PoWConfig {
	return c.config
}

func (c *PoWDeviceMng) Init() error {
	// allocate all exec
	c.threads = make([]itfcs.PoWThread, 0)
	wkrch := c.alloter.Allocate()
	for {
		//fmt.Println("(c *PoWDeviceMng) Init()-------------------------------")
		exec := <-wkrch
		if exec == nil {
			break
		}
		thr := NewPoWThreadMng(exec)
		c.threads = append(c.threads, thr)
	}
	// call init
	for _, thr := range c.threads {
		e := thr.Init()
		if e != nil {
			return e
		}
	}
	return nil
}

func (c *PoWDeviceMng) StopMining() {
	for _, thr := range c.threads {
		thr.StopMining()
	}
}

// find block
func (c *PoWDeviceMng) DoMining(stopmark *byte, inputCh chan *itfcs.PoWStuffBriefData) (*itfcs.PoWResultData, error) {

	var execNum = len(c.threads)
	var resChs = make(chan *itfcs.PoWResultData, execNum+1)
	var execWait = sync.WaitGroup{}
	execWait.Add(1)

	// final result
	var most_result *itfcs.PoWResultData = nil
	var brief_ccl = <-inputCh
	if brief_ccl == nil {
		return nil, fmt.Errorf("Error: Cannot read PoWStuffBriefData")
	}

	// show
	var block_height = brief_ccl.BlockHeadMeta.GetHeight()
	var tar_diff_str = hex.EncodeToString(difficulty.DifficultyUint32ToHash(brief_ccl.BlockHeadMeta.GetDifficulty()))
	tar_diff_str = strings.TrimRight(tar_diff_str, "0")
	exec_start_time := time.Now()
	fmt.Printf("[%s] do mining: ‹%d› thr: %s",
		time.Now().Format("01/02 15:04:05"),
		block_height, tar_diff_str)
	if c.config.IsDetailLog() {
		fmt.Print("... ")
	} else {
		fmt.Print("      ")
	}

	var target_hash = difficulty.DifficultyUint32ToHash(brief_ccl.BlockHeadMeta.GetDifficulty())

	for i := 0; i < execNum; i++ {
		go func(idx int, target_hash fields.Hash) {
			// do mining
			var exec = c.threads[idx]
			var brief = <-inputCh
			if brief == nil {
				resChs <- nil
				fmt.Println("exec.Run cannot read PoWStuffBriefData")
				return
			}
			e := exec.DoMining(stopmark, target_hash, *brief, resChs)
			if e != nil {
				resChs <- nil
				fmt.Println("exec.DoMining error: ", e)
				return
			}
		}(i, target_hash)
	}

	// deal result
	go func() {
		for i := 0; i < execNum; i++ {
			var res = <-resChs
			if res == nil {
				continue
			}
			if res.FindSuccess.Check() {
				// find block
				*stopmark = 1
				c.StopMining()
				fmt.Printf("[%s] upload success find block: <%d> hash: %s\n",
					time.Now().Format("01/02 15:04:05"),
					res.BlockHeight,
					res.ResultHash.ToHex(),
				)
				most_result = res
				continue
			}
			if most_result == nil {
				most_result = res
				continue
			}
			if bytes.Compare(most_result.ResultHash, res.ResultHash) == 1 {
				most_result = res
			}
			// next res
		}
		// all item ok
		execWait.Done()
	}()

	// wait all exec down
	execWait.Wait()

	if most_result != nil && !most_result.FindSuccess.Check() {
		// upload hash
		digg_time := time.Since(exec_start_time).Seconds()
		var lphr = difficulty.ConvertHashToRate(block_height, most_result.ResultHash, int64(digg_time))
		var lphr_show = difficulty.ConvertPowPowerToShowFormat(lphr)

		// count total hr
		hxrate_show_count++
		if hxrate_show_ttvalue == nil {
			hxrate_show_ttvalue = lphr
		} else {
			hxrate_show_ttvalue = hxrate_show_ttvalue.Add(hxrate_show_ttvalue, lphr)
		}
		var lphr_average = difficulty.ConvertPowPowerToShowFormat(big.NewInt(0).Div(hxrate_show_ttvalue, big.NewInt(hxrate_show_count)))

		fmt.Printf("upload power: %s... chr: %s hashrate: %s\n",
			most_result.ResultHash.ToHex()[0:24],
			lphr_show, lphr_average,
		)
	}

	// clean
	close(resChs)

	// ok ret
	return most_result, nil
}
