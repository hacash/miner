package device

import (
	"bytes"
	"fmt"
	"github.com/hacash/core/fields"
	itfcs "github.com/hacash/miner/interfaces"
	"github.com/hacash/mint/difficulty"
	"sync"
	"time"
)

type PoWDeviceMng struct {
	alloter itfcs.PoWExecute
	threads []itfcs.PoWThread
}

func NewPoWDeviceMng(alloter itfcs.PoWExecute) *PoWDeviceMng {
	return &PoWDeviceMng{
		alloter: alloter,
		threads: make([]itfcs.PoWThread, 0),
	}
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
	var resChs = make(chan *itfcs.PoWResultData, execNum)
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
	exec_start_time := time.Now()
	fmt.Printf("device mining: ‹%d›...", block_height)

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

	if most_result != nil {
		digg_time := time.Since(exec_start_time).Seconds()
		fmt.Printf("power: %s... hashrate: %s\n", most_result.ResultHash.ToHex()[0:16],
			difficulty.ConvertHashToRateShow(block_height, most_result.ResultHash, int64(digg_time)))
	}

	// ok ret
	return most_result, nil
}
