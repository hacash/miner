package memtxpool

func (p *MemTxPool) loop() {

	for {
		select {
		// diamond create event handler
		case <-p.newDiamondCreateCh:
			p.changeLock.Lock()
			p.diamondCreateTxGroup.Clean() // delete all diamond create tx
			p.changeLock.Unlock()
		}

	}

}
