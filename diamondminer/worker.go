package diamondminer

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/stores"
	"github.com/hacash/x16rs"
	"sync"
)

func (d *DiamondMiner) RunMining(prevDiamond *stores.DiamondSmelt, diamondCreateActionCh chan *actions.Action_4_DiamondCreate) {
	d.changeLock.Lock()
	defer d.changeLock.Unlock()

	fmt.Printf("do diamond mining... number: %d, supervene: %d, start worker:", prevDiamond.Number+1, d.Config.Supervene)

	// stop prev all
	for _, v := range d.stopMarks {
		*v = 1 // stop
	}

	var stopMark byte = 0
	d.stopMarks[&stopMark] = &stopMark

	// do mining
	go func(supervene int, stopMark *byte, prevDiamond *stores.DiamondSmelt, diamondCreateActionCh chan *actions.Action_4_DiamondCreate) {

		var current_i uint32 = 0
		var current_lock = sync.Mutex{}

		for i := 0; i < supervene; i++ {
			go func(i int) {
			NEXTMINING:
				var my_i uint32 = 0
				current_lock.Lock()
				current_i++
				my_i = current_i
				current_lock.Unlock()
				// call mining
				tarnumber := int(prevDiamond.Number) + 1
				retExtMsg := make([]byte, 32) // 随机字段值，让同一个地址配置也可以挖不同的的钻石
				mnstart, mnend := my_i, my_i+1
				if uint32(tarnumber) > actions.DiamondCreateCustomMessageAboveNumber {
					mnstart, mnend = 0, 4294967290
					rand.Read(retExtMsg)
				}
				fmt.Printf(" #%d", my_i)
				retNonce, diamondFullStr := x16rs.MinerHacashDiamond(mnstart, mnend, tarnumber, stopMark, prevDiamond.ContainBlockHash, d.Config.Rewards, retExtMsg)
				retNonceNum := binary.BigEndian.Uint64(retNonce)
				if retNonceNum > 0 {
					fmt.Printf("\n\n[Diamond Miner] Success find a diamond: <%s>, number: %d, nonce: %d .\n\n", diamondFullStr, tarnumber, retNonceNum)
					// success
					diamondCreateActionCh <- parsediamondCreateAction(diamondFullStr, prevDiamond, retNonce, d.Config.Rewards, retExtMsg)
					// go to next loop
				}

				if *stopMark == 1 {
					return // set stop
				}
				// LOOP NEXT
				goto NEXTMINING
			}(i)
		}

	}(d.Config.Supervene, &stopMark, prevDiamond, diamondCreateActionCh)

}

func parsediamondCreateAction(
	diamondFullStr string,
	prevDiamond *stores.DiamondSmelt,
	retNonce []byte,
	rewards fields.Address,
	extMsg []byte,
) *actions.Action_4_DiamondCreate {
	newact := &actions.Action_4_DiamondCreate{
		Diamond:       []byte(diamondFullStr)[10:16],
		Number:        prevDiamond.Number + 1,
		PrevHash:      prevDiamond.ContainBlockHash,
		Nonce:         retNonce,
		Address:       rewards,
		CustomMessage: extMsg,
	}
	return newact
}
