package memtxpool

func (p *MemTxPool) loop() {

	for {
		select {
		// diamond create event handler
		case <-p.newDiamondCreateCh:
			// fmt.Println("p.newDiamondCreateCh  p.diamondCreateTxGroup.Clean()", )
			p.changeLock.Lock()
			p.diamondCreateTxGroup.Clean() // delete all diamond create tx
			p.changeLock.Unlock()
		}

	}

}
