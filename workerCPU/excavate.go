package workerCPU

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/mint/difficulty"
	"math/big"
	"sync"
	"sync/atomic"
	"time"
)

func (c *CPUWorker) Excavate(miningStuffCh chan interfaces.PowWorkerMiningStuffItem, resultCh chan interfaces.PowWorkerMiningStuffItem) {
	c.miningStuffCh = miningStuffCh
	c.resultCh = resultCh
}

func (c *CPUWorker) ExcavateOld(miningStuffCh chan interfaces.PowWorkerMiningStuffItem, resultCh chan interfaces.PowWorkerMiningStuffItem) {

	supervene := uint32(c.config.Supervene) // 并发线程
	// -------- test start --------
	//maxuint32 = uint32(2294900)
	//supervene := uint32(6)
	// -------- test end   --------
	//nonceSpace := maxuint32 / supervene

	var currentDoMiningHeight uint64 = 0
	// 时间统计
	timestart := time.Now()

	// 永远挖掘
	for {

		if c.pendingBlockHeight == 0 {
			// 还没准备好
			time.Sleep(time.Millisecond * 5)
			continue
		}

		var successFindBlock = false
		var mostPowerHash []byte = nil
		var mostPowerStuffItem interfaces.PowWorkerMiningStuffItem = nil

		// new run
		var nextstop byte = 0
		c.nextMarks.Store(&nextstop, &nextstop)

		var syncWait = sync.WaitGroup{}
		syncWait.Add(int(supervene))

		//fmt.Println("Excavate",  powmsg.CoinbaseMsgNum, powmsg.BlockHeadMeta)

		// 并发挖掘
		for i := uint32(0); i < supervene; i++ {
		REGETSTUFF:
			stuffitem := <-miningStuffCh // 取出挖掘素材
			tarblock := stuffitem.GetHeadMetaBlock()
			tarheight := tarblock.GetHeight()
			if c.pendingBlockHeight != tarheight {
				goto REGETSTUFF // 放弃挖掘高度不匹配的区块
			}
			// 打印
			if currentDoMiningHeight != c.pendingBlockHeight {
				if !c.config.IsReportPower {
					fmt.Print("\n")
				}
				showtarget := c.pendingBlockHeight%10 == 0
				if showtarget {
					diffhex := hex.EncodeToString(difficulty.Uint32ToHash(tarheight, tarblock.GetDifficulty()))
					fmt.Print("do mining height:‹", c.pendingBlockHeight, "›, target: ", diffhex[0:32], "...")
				} else {
					fmt.Print("do mining height:‹", c.pendingBlockHeight, "›...")
				}
				currentDoMiningHeight = c.pendingBlockHeight
				timestart = time.Now()
			} else {
				fmt.Print(".")
			}
			// 启动挖掘线程
			go func(stuffitem interfaces.PowWorkerMiningStuffItem, tarblock interfaces.Block, tarheight uint64) {
				defer syncWait.Done()
				// 执行
				//fmt.Println("c.runOne(stuffitem)")
				result := c.runOne(stuffitem, &nextstop)
				if result != nil {
					stuffitem.SetHeadNonce(result.nonceBytes)      // 算力统计
					stuffitem.SetMiningSuccessed(result.isSuccess) // 设置挖矿状态
					if result.isSuccess && atomic.CompareAndSwapUint32(c.successMiningMark, 0, 1) {
						successFindBlock = true
						c.pendingBlockHeight = 0 // 暂时不挖掘了
						// 成功
						resultCh <- stuffitem // 成功返回
						fmt.Printf("found.\n[⬤◆◆] Successfully minted a block height: %d, hash: %s, nonce: %s. \n",
							stuffitem.GetHeadMetaBlock().GetHeight(),
							hex.EncodeToString(result.powerHash),
							hex.EncodeToString(result.nonceBytes))
						if c.pendingBlockHeight < 10000 {
							//time.Sleep(time.Second) // 暂停一秒钟等待挖掘下一个区块
						}
					} else if c.config.IsReportPower {
						// 算力统计
						if mostPowerStuffItem == nil || difficulty.CheckHashDifficultySatisfy(result.powerHash, mostPowerHash) {
							mostPowerHash = result.powerHash
							mostPowerStuffItem = stuffitem
						}
					}
				}
			}(stuffitem, tarblock, tarheight)
		}

		// wait
		syncWait.Wait()

		c.nextMarks.Delete(&nextstop) // clean

		// 上报算力
		if !successFindBlock {
			if c.config.IsReportPower && mostPowerHash != nil && c.pendingBlockHeight == mostPowerStuffItem.GetHeadMetaBlock().GetHeight()+1 {
				// 上报算力
				usetimesec := time.Now().Unix() - timestart.Unix()
				hxworth := difficulty.CalculateHashWorth(mostPowerStuffItem.GetHeadMetaBlock().GetHeight(), mostPowerHash)
				hashrate := new(big.Int).Div(hxworth, big.NewInt(usetimesec))
				hashrateshow := difficulty.ConvertPowPowerToShowFormat(hashrate)
				fmt.Printf("upload power: %s, time: %ds, hashrate: %s.\n",
					hex.EncodeToString(mostPowerHash[0:16]),
					usetimesec, hashrateshow,
				)
				// 返回算力统计
				resultCh <- mostPowerStuffItem
				mostPowerStuffItem = nil
				mostPowerHash = nil
			}
		}

	}
}
