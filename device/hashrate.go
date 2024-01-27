package device

import (
	"bytes"
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/mint"
	"github.com/hacash/mint/coinbase"
	"github.com/hacash/mint/difficulty"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type HashrateLogTable struct {
	wides []int
	logts []string
	hxsch chan fields.Hash

	hsrtc *time.Ticker
	lsnum int

	httl1 []string
	httl2 []string
	linex string

	blktm                 int64
	current_blk           interfaces.Block
	blk_rate              *big.Int
	period_max_power_hash fields.Hash

	hxrate_show_count   int64
	hxrate_show_ttvalue *big.Int

	cache_head string
}

func NewHashrateLog() *HashrateLogTable {
	var log = &HashrateLogTable{
		wides: []int{8, 9, 16, 10, 11, 12},
		logts: []string{"*", "*", "*", "*", "*", "Wait..."},
		httl1: []string{"Mining", "Mainnet", "Upload", "Hashrate", "Estimate", "Native"},
		httl2: []string{"Height", "Hashrate", "Power Hash", "Percent", "HAC/Day", "Hashrate"},
		//
		linex: "│",
		blktm: int64(mint.EachBlockRequiredTargetTime),
		hxsch: make(chan fields.Hash, 4),
		lsnum: 0,
	}
	var maxhash = bytes.Repeat([]byte{255}, 32)
	log.blk_rate = big.NewInt(0).SetBytes(maxhash)
	log.hsrtc = time.NewTicker(time.Second * time.Duration(log.blktm))
	log.period_max_power_hash = maxhash
	log.hxrate_show_ttvalue = big.NewInt(0)
	go log.loopRecordHash()
	go log.loopPrintHashrate()
	// ok
	return log
}

func (h *HashrateLogTable) loopPrintHashrate() {
	for {
		<-h.hsrtc.C
		var hei = h.current_blk.GetHeight()
		var lphr = difficulty.ConvertHashToRate(hei, h.period_max_power_hash, h.blktm)
		h.hxrate_show_count++
		h.hxrate_show_ttvalue.Add(h.hxrate_show_ttvalue, lphr)
		var average_lphr = big.NewInt(0).Div(h.hxrate_show_ttvalue, big.NewInt(h.hxrate_show_count))
		var average_hashrate = difficulty.ConvertPowPowerToShowFormat(average_lphr)
		// show
		h.logts[5] = average_hashrate
		var ratio = big.NewFloat(float64(100))
		ratio = ratio.Mul(ratio, big.NewFloat(0).SetInt(average_lphr))
		ratio = ratio.Quo(ratio, big.NewFloat(0).SetInt(h.blk_rate))
		percentf, _ := ratio.Float64()
		h.logts[3] = fmt.Sprintf("%.6f%%", percentf)
		var getcoinday = big.NewInt(288)
		getcoinday = getcoinday.Mul(getcoinday, big.NewInt(int64(coinbase.BlockCoinBaseRewardNumber(hei))))
		getcoindayf := big.NewFloat(0).SetInt(getcoinday)
		getcoindayf = getcoindayf.Quo(getcoindayf, big.NewFloat(100))
		getcoindayf = getcoindayf.Mul(getcoindayf, ratio)
		gcdf, _ := getcoindayf.Float64()
		h.logts[4] = fmt.Sprintf("%.8f", gcdf)
		h.flushLine() // do print
		// reset
		h.period_max_power_hash = bytes.Repeat([]byte{255}, 32)
		// next print
	}
}

func (h *HashrateLogTable) loopRecordHash() {
	for {
		hx := <-h.hxsch
		if bytes.Compare(hx, h.period_max_power_hash) == -1 {
			h.period_max_power_hash = hx
			h.logts[2] = hx.ToHex()[0:20]
			// for print hashrate
			//fmt.Println(hx.ToHex())
		}
	}
}

func (h *HashrateLogTable) RecordHashChan() chan fields.Hash {
	return h.hxsch
}

func (h *HashrateLogTable) UpdateMiningBlock(blk interfaces.Block) {
	h.current_blk = blk
	var hei = blk.GetHeight()
	h.logts[0] = fmt.Sprintf("%d", hei)
	//
	targetHashrate := difficulty.ConvertHashToRate(hei,
		difficulty.Uint32ToHash(hei, blk.GetDifficulty()), int64(mint.EachBlockRequiredTargetTime))
	h.blk_rate = targetHashrate
	targetHashrateShow := difficulty.ConvertPowPowerToShowFormat(targetHashrate)
	h.logts[1] = strings.TrimRight(targetHashrateShow, "H/s")
	// ok print
	h.flushLine()
}

func (h *HashrateLogTable) flushLine() {
	var lines string = ""
	if h.lsnum%50 == 0 {
		lines = h.sprinthead() + "\n"
	}
	// line
	lines += h.dealline(h.logts)
	// ok print
	fmt.Println(lines)
	h.lsnum++
}

func (h *HashrateLogTable) dealline(strs []string) string {
	return h.sprintline(strs, h.linex)
}

func (h *HashrateLogTable) sprintline(strs []string, spx string) string {
	l1 := len(strs)
	l2 := len(h.wides)
	lu := l1
	if l2 < l1 {
		lu = l2
	}
	var res = make([]string, 1, lu+2)
	res[0] = ""
	for i := 0; i < lu; i++ {
		w := strconv.Itoa(h.wides[i])
		res = append(res, fmt.Sprintf("%"+w+"."+w+"s", strs[i]))
	}
	res = append(res, "")
	return strings.Join(res, spx)
}

func (h *HashrateLogTable) sprinthead() string {
	if len(h.cache_head) > 0 {
		return h.cache_head
	}

	w := len(h.wides)
	hdl := strings.Repeat("─", 100)
	hdls := make([]string, 0, w)
	for i := 0; i < w; i++ {
		hdls = append(hdls, hdl)
	}
	// line
	var heads = make([]string, 0, 3)
	heads = append(heads, h.sprintline(hdls, "┼"))
	heads = append(heads, h.dealline(h.httl1))
	heads = append(heads, h.dealline(h.httl2))
	// ok
	h.cache_head = strings.Join(heads, "\n")
	return h.cache_head
}
