package minerworkerwrap

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/mint/difficulty"
	"strings"
	"time"
)

// to do next
func (g *WorkerWrap) DoNextMining(pendingHeight uint64) {

	// All excavation before stopping
	g.StopAllMining()
	//fmt.Println("g.StopAllMining()")

	// stop mark
	var stopmark byte = 0
	g.stopMarks.Store(&stopmark, &stopmark)
	defer g.stopMarks.Delete(&stopmark)

	// Serial synchronization
	g.stepLock.Lock()
	defer g.stepLock.Unlock()

	//fmt.Println("STARTDOMINING....")

	// Test report
	//go func() {
	//	time.Sleep(time.Second * 10)
	//	stopmark = 1
	//}()

	// Time statistics
	timestart := time.Now()
	hasshowinfo := false

	// Circular mining
STARTDOMINING:

	if stopmark == 1 {
		return // Stop current mining
	}
	// Take out the stuff until it is consistent with the mining target
	// Concurrent quantity
	supervene := g.powdevice.GetSuperveneWide()
	stuffitemlist := make([]interfaces.PowWorkerMiningStuffItem, supervene)
	blockheadmetasary := make([][]byte, supervene)
	oksuffnum := 0
	var stuffitem interfaces.PowWorkerMiningStuffItem = nil
	for {
		if stopmark == 1 {
			return // Stop current mining
		}
		stuffitem = <-g.miningStuffCh
		tarblock := stuffitem.GetHeadMetaBlock()
		tarheight := tarblock.GetHeight()
		if tarheight != pendingHeight {
			time.Sleep(time.Millisecond * 100)
			continue // Waiting for next acquisition
		}
		//fmt.Println(tarblock.GetMrklRoot())
		blockheadmeatastuff := blocks.CalculateBlockHashBaseStuff(tarblock)
		stuffitemlist[oksuffnum] = stuffitem
		blockheadmetasary[oksuffnum] = blockheadmeatastuff // block mining stuff
		oksuffnum++
		if oksuffnum == supervene {
			break // Start digging
		}
	}

	// Print
	if hasshowinfo == false {
		hasshowinfo = true
		diffhex := hex.EncodeToString(difficulty.Uint32ToHash(pendingHeight, stuffitem.GetHeadMetaBlock().GetDifficulty()))
		fmt.Print("do mining height:‹", pendingHeight, "›, difficulty: ", strings.TrimRight(diffhex, "0"), "..")
	}

	// Start digging
	pendingblock := stuffitem.GetHeadMetaBlock()
	tardiffhash := difficulty.Uint32ToHash(pendingHeight, pendingblock.GetDifficulty())
	if stopmark == 1 {
		return // Stop current mining
	}
	// do mining
	//fmt.Println(supervene, blockheadmetasary)
	fmt.Print(".")
	success, endstuffidx, nonce, endhash := g.powdevice.DoMining(pendingHeight, g.config.IsReportHashrate, &stopmark, tardiffhash, blockheadmetasary)

	// Returned stuff
	endstuffitem := stuffitemlist[endstuffidx]

	// Judge whether the mining is successful
	if success {
		endstuffitem.SetMiningSuccessed(true)
		endstuffitem.SetHeadNonce(nonce)
		// Mining success Report
		g.resultCh <- endstuffitem
		// Print
		fmt.Printf("found success.\n[⬤◆◆] Successfully mined a block height: %d, hash: %s, nonce: %s, time: %s. \n",
			endstuffitem.GetHeadMetaBlock().GetHeight(),
			hex.EncodeToString(endhash),
			hex.EncodeToString(nonce),
			time.Now().Format("01/02 15:04:05"),
		)
		// Success and return
		return
	}

	// Judge whether to stop
	if stopmark == 0 {
		// Not stopped, next excavation
		goto STARTDOMINING
	}

	// 上报算力 // 检查 nonce
	if g.config.IsReportHashrate && nonce != nil && len(nonce) == 4 {
		endstuffitem.SetMiningSuccessed(false)
		endstuffitem.SetHeadNonce(nonce)
		// Reported computing power
		g.resultCh <- endstuffitem
		// Print
		usetimesec := time.Now().Unix() - timestart.Unix()
		hashrateshow := difficulty.ConvertHashToRateShow(pendingHeight, endhash, usetimesec)
		fmt.Printf("upload power: %s, time: %ds, hashrate: %s.\n",
			hex.EncodeToString(endhash[0:16]),
			usetimesec, hashrateshow,
		)
	} else {
		fmt.Printf("ok.\n")
	}
}
