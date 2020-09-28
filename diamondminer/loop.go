package diamondminer

import "time"

func (d *DiamondMiner) loop() {

	var autobidTimeout = time.NewTicker(time.Second * 10)

	for {
		select {
		case newdiamond := <-d.newDiamondBeFoundCh:
			d.RunMining(newdiamond, d.successMiningDiamondCh)

		case findsuccess := <-d.successMiningDiamondCh:
			go d.successFindDiamondAddTxPool(findsuccess)

		case <-autobidTimeout.C:
			if d.Config.AutoBid {
				go d.doAutoBidForMyDiamond()
			}
		}
	}
}
