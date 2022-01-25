package diamondminer

import "time"

func (d *DiamondMiner) loop() {

	cktm := time.Duration(d.Config.AutoCheckInterval*1000) * time.Millisecond
	var autobidInterval = time.NewTicker(cktm)

	for {
		select {
		case newdiamond := <-d.newDiamondBeFoundCh:
			d.RunMining(newdiamond, d.successMiningDiamondCh)

		case findsuccess := <-d.successMiningDiamondCh:
			go d.successFindDiamondAddTxPool(findsuccess)

		case <-autobidInterval.C:
			if d.Config.AutoBid {
				d.doAutoBidForMyDiamond()
			}
		}
	}
}
