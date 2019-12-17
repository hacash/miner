package memtxpool

import (
	"github.com/hacash/core/interfaces"
)

func (p *MemTxPool) RemoveTxs(txs []interfaces.Transaction) {
	p.changeLock.Lock()
	defer p.changeLock.Unlock()
	// remove
	for _, tx := range txs {
		txhx := tx.Hash()
		p.diamondCreateTxGroup.RemoveByTxHash(txhx)
		if hav := p.simpleTxGroup.RemoveByTxHash(txhx); hav != nil {
			// sub count
			p.txTotalCount -= 1
			p.txTotalSize -= uint64(hav.size)
		}
	}

}
