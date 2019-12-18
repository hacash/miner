package miningpool

func (p *MinerPool) loop() {

	for {
		select {

		case arriveNewBlock := <-p.newBlockOnInsertFeedCh:

		}
	}

}
