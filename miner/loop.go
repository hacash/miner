package miner

import (
	"github.com/hacash/core/interfaces"
	"time"
)

func (m *Miner) loop() {

	for {
		select {

		case newblk := <-m.newBlockOnInsertCh:
			mark := newblk.OriginMark()
			if mark == "discover" || mark == "mining" {
				m.StopMining()
				time.Sleep(time.Millisecond * 10)
				m.txpool.RemoveTxs(newblk.(interfaces.Block).GetTrsList())
				m.StartMining()
				// restart mining
			}
		}

	}

}
