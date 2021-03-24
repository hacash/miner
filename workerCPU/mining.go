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

// stop mining
func (c *CPUWorker) NextMining(pendingHeight uint64) {

	//fmt.Println("pendingHeight:", pendingHeight)

	// 结束所有之前的挖矿
	c.nextMarks.Range(func(k interface{}, v interface{}) bool {
		mk := v.(*byte)
		*mk = 1 // set stop
		return false
	})

	// 流线锁
	c.miningstreamlock.Lock()
	defer c.miningstreamlock.Unlock()

	var smm uint32 = 0
	successMiningMark := &smm // 未成功标记

	// 开始新的挖款
	supervene := int(c.config.Supervene) // 并发线程
	// 取出 stuff 直到于挖掘目标一致
	for {
		firststuffitem := <-c.miningStuffCh
		tarblock := firststuffitem.GetHeadMetaBlock()
		tarheight := tarblock.GetHeight()
		if tarheight != pendingHeight {
			continue
		}
		showtarget := pendingHeight%10 == 0
		if showtarget {
			diffhex := hex.EncodeToString(difficulty.Uint32ToHash(tarheight, tarblock.GetDifficulty()))
			fmt.Print("do mining height:‹", pendingHeight, "›, target: ", diffhex[0:32], "...")
		} else {
			fmt.Print("do mining height:‹", pendingHeight, "›...")
		}
		break // 开始挖掘
	}
	// 算力统计

	var successFindResult *miningBlockReturn = nil
	var successFindBlockStuff interfaces.PowWorkerMiningStuffItem = nil // 挖掘成功
	var mostPowerHash []byte = nil
	var mostPowerStuffItem interfaces.PowWorkerMiningStuffItem = nil

	// stop mark
	var nextstop byte = 0
	c.nextMarks.Store(&nextstop, &nextstop)
	// group
	var syncWait = sync.WaitGroup{}
	syncWait.Add(supervene)
	// 时间统计
	timestart := time.Now()
	for i := 0; i < supervene; i++ {
		go func(i int, nextstop *byte) {
			defer syncWait.Done()
			for {
				// 获取一个挖矿素材
				stuffitem := <-c.miningStuffCh
				tarblock := stuffitem.GetHeadMetaBlock()
				tarheight := tarblock.GetHeight()
				if tarheight != pendingHeight {
					continue
				}
				// 开始挖掘
				fmt.Print(".")
				result := c.runOne(stuffitem, nextstop)
				if result != nil {
					stuffitem.SetHeadNonce(result.nonceBytes)      // 算力统计
					stuffitem.SetMiningSuccessed(result.isSuccess) // 设置挖矿状态
					if result.isSuccess && atomic.CompareAndSwapUint32(successMiningMark, 0, 1) {
						successFindResult = result
						successFindBlockStuff = stuffitem
						break // 本次挖矿停止
					} else if c.config.IsReportPower {
						// 算力统计
						if mostPowerStuffItem == nil || difficulty.CheckHashDifficultySatisfy(result.powerHash, mostPowerHash) {
							mostPowerHash = result.powerHash
							mostPowerStuffItem = stuffitem
						}
					}
				}
				if *nextstop == 1 {
					break // 本次挖矿停止
				}
			}
		}(i, &nextstop)
	}

	syncWait.Wait()

	// 本次挖矿全部完成
	c.nextMarks.Delete(&nextstop) // clean

	// 上报算力
	if successFindBlockStuff == nil {
		if c.config.IsReportPower && mostPowerHash != nil {
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
			c.resultCh <- mostPowerStuffItem
			mostPowerStuffItem = nil
			mostPowerHash = nil
		} else {
			fmt.Printf("ok.\n")
		}
	} else {

		// 成功
		c.resultCh <- successFindBlockStuff // 成功返回
		fmt.Printf("found success.\n[⬤◆◆] Successfully minted a block height: %d, hash: %s, nonce: %s. \n",
			successFindBlockStuff.GetHeadMetaBlock().GetHeight(),
			hex.EncodeToString(successFindResult.powerHash),
			hex.EncodeToString(successFindResult.nonceBytes))
	}

}
