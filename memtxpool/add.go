package memtxpool

import (
	"github.com/hacash/core/actions"
	"github.com/hacash/core/interfaces"
)

func (p *MemTxPool) SubmitTx(tx interfaces.Transaction) {
	p.changeLock.Lock()
	defer p.changeLock.Unlock()

	// is diamond create tx and not check
	for _, act := range tx.GetActions() {
		if _, ok := act.(*actions.Action_4_DiamondCreate); ok {
			p.addTx(tx, p.diamondCreateTxs)
			return
		}
	}

	// simple tx and do check
	diamond, err := p.blockchain.ValidateTransaction(tx)
	if err != nil {
		return // error tx
	}

}

func (p *MemTxPool) addTx(tx interfaces.Transaction, list *TxItem) {

}
