package miner

import "time"

func (m *Miner) loop() {

	for {
		select {

		case newblk := <-m.newBlockOnInsertCh:
			mark := newblk.OriginMark()
			m.txpool.RemoveTxs(newblk.GetTransactions())
			if mark == "discover" || mark == "mining" {
				m.StopMining()
				time.Sleep(time.Millisecond * 10)
				m.StartMining()
				// restart mining
			}
		}

	}

}
