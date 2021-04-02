package minerworkerwrap

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/mint/difficulty"
	"math/big"
	"strings"
	"time"
)

// to do next
func (g *WorkerWrap) DoNextMining(pendingHeight uint64) {

	// 停止之前所有挖款
	g.StopAllMining()
	//fmt.Println("g.StopAllMining()")

	// stop mark
	var stopmark byte = 0
	g.stopMarks.Store(&stopmark, &stopmark)
	defer g.stopMarks.Delete(&stopmark)

	// 串行同步
	g.stepLock.Lock()
	defer g.stepLock.Unlock()

	//fmt.Println("STARTDOMINING....")

	// 时间统计
	timestart := time.Now()
	hasshowinfo := false

	// 循环挖矿
STARTDOMINING:

	if stopmark == 1 {
		return // 停止当前挖矿
	}
	// 取出 stuff 直到于挖掘目标一致
	// 并发数量
	supervene := g.powdevice.GetSuperveneWide()
	stuffitemlist := make([]interfaces.PowWorkerMiningStuffItem, supervene)
	blockheadmetasary := make([][]byte, supervene)
	oksuffnum := 0
	var stuffitem interfaces.PowWorkerMiningStuffItem = nil
	for {
		if stopmark == 1 {
			return // 停止当前挖矿
		}
		stuffitem = <-g.miningStuffCh
		tarblock := stuffitem.GetHeadMetaBlock()
		tarheight := tarblock.GetHeight()
		if tarheight != pendingHeight {
			time.Sleep(time.Millisecond * 100)
			continue // 等待下一次获取
		}
		//fmt.Println(tarblock.GetMrklRoot())
		blockheadmeatastuff := blocks.CalculateBlockHashBaseStuff(tarblock)
		stuffitemlist[oksuffnum] = stuffitem
		blockheadmetasary[oksuffnum] = blockheadmeatastuff // block mining stuff
		oksuffnum++
		if oksuffnum == supervene {
			break // 开始挖掘
		}
	}

	// 打印
	if hasshowinfo == false {
		hasshowinfo = true
		diffhex := hex.EncodeToString(difficulty.Uint32ToHash(pendingHeight, stuffitem.GetHeadMetaBlock().GetDifficulty()))
		fmt.Print("do mining height:‹", pendingHeight, "›, difficulty: ", strings.TrimRight(diffhex, "0"), "..")
	}

	// 开始挖掘
	pendingblock := stuffitem.GetHeadMetaBlock()
	tardiffhash := difficulty.Uint32ToHash(pendingHeight, pendingblock.GetDifficulty())
	if stopmark == 1 {
		return // 停止当前挖矿
	}
	// do mining
	//fmt.Println(supervene, blockheadmetasary)
	fmt.Print(".")
	success, endstuffidx, nonce, endhash := g.powdevice.DoMining(pendingHeight, g.config.IsReportHashrate, &stopmark, tardiffhash, blockheadmetasary)

	// 返回的 stuff
	endstuffitem := stuffitemlist[endstuffidx]

	// 判断是否挖矿成功
	if success {
		endstuffitem.SetMiningSuccessed(true)
		endstuffitem.SetHeadNonce(nonce)
		// 挖矿成功上报
		g.resultCh <- endstuffitem
		// 打印
		fmt.Printf("found success.\n[⬤◆◆] Successfully minted a block height: %d, hash: %s, nonce: %s. \n",
			endstuffitem.GetHeadMetaBlock().GetHeight(),
			hex.EncodeToString(endhash),
			hex.EncodeToString(nonce))
		// 成功并返回
		return
	}

	// 判断是否停止
	if stopmark == 0 {
		// 未停止，进行下一次挖款
		goto STARTDOMINING
	}

	// 上报算力 // 检查 nonce
	if g.config.IsReportHashrate && nonce != nil && len(nonce) == 4 {
		endstuffitem.SetHeadNonce(nonce)
		// 上报算力
		g.resultCh <- endstuffitem
		// 打印
		usetimesec := time.Now().Unix() - timestart.Unix()
		hxworth := difficulty.CalculateHashWorth(endstuffitem.GetHeadMetaBlock().GetHeight(), endhash)
		hashrate := new(big.Int).Div(hxworth, big.NewInt(usetimesec))
		hashrateshow := difficulty.ConvertPowPowerToShowFormat(hashrate)
		fmt.Printf("upload power: %s, time: %ds, hashrate: %s.\n",
			hex.EncodeToString(endhash[0:16]),
			usetimesec, hashrateshow,
		)
	} else {
		fmt.Printf("ok.\n")
	}
}
