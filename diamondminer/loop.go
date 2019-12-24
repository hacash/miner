package diamondminer

func (d *DiamondMiner) loop() {
	for {
		select {
		case newdiamond := <-d.newDiamondBeFoundCh:
			d.RunMining(newdiamond, d.successMiningDiamondCh)

		case findsuccess := <-d.successMiningDiamondCh:
			go d.successFindDiamondAddTxPool(findsuccess)

		}
	}
}
