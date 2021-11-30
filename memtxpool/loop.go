package memtxpool

import (
	"github.com/hacash/core/interfacev2"
	"time"
)

func (p *MemTxPool) loop() {

	automaticallyCleanInvalidTransactionsTicker := time.NewTicker(time.Minute * 9)

	for {
		select {
		// diamond create event handler
		case <-p.newDiamondCreateCh:
			// fmt.Println("p.newDiamondCreateCh  p.diamondCreateTxGroup.Clean()", )
			p.changeLock.Lock()
			p.diamondCreateTxGroup.Clean() // delete all diamond create tx
			p.changeLock.Unlock()

		case newblk := <-p.newBlockOnInsertCh:
			p.RemoveTxs(newblk.GetTransactions())
			txs := p.removeTxsOnNextBlockArrive
			p.removeTxsOnNextBlockArrive = []interfacev2.Transaction{}
			p.RemoveTxs(txs)

		case <-automaticallyCleanInvalidTransactionsTicker.C:
			if p.automaticallyCleanInvalidTransactions {
				p.doCleanInvalidTransactions()
			}

		}
	}

}
