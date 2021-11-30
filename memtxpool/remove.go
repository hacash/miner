package memtxpool

import (
	"github.com/hacash/core/interfacev2"
	"github.com/hacash/mint"
)

func (p *MemTxPool) RemoveTxs(txs []interfacev2.Transaction) {
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

func (p *MemTxPool) RemoveTxsOnNextBlockArrive(txs []interfacev2.Transaction) {
	p.changeLock.Lock()
	defer p.changeLock.Unlock()

	if txs != nil && len(txs) > 0 {
		p.removeTxsOnNextBlockArrive = append(p.removeTxsOnNextBlockArrive, txs...)
	}
}

func (p *MemTxPool) doCleanInvalidTransactions() {
	p.changeLock.Lock()
	defer p.changeLock.Unlock()

	/*
		tempState, err := p.blockchain.State().Fork()
		if err != nil {
			return
		}
		defer tempState.Destory()
	*/

	sizeCount := uint32(0)
	head := p.simpleTxGroup.Head
	for {
		if head == nil {
			break
		}
		if sizeCount > mint.SingleBlockMaxSize*2 {
			break
		}
		sizeCount += head.size
		// check
		e2 := p.blockchain.ValidateTransactionForTxPool(head.tx)
		if e2 != nil {
			p.removeTxsOnNextBlockArrive = append(p.removeTxsOnNextBlockArrive, head.tx)
		}
		/*
			txState, e1 := tempState.Fork()
			if e1 != nil {
				return
			}
			e2 := head.tx.WriteinChainState(txState)
			if e2 != nil {
				p.removeTxsOnNextBlockArrive = append(p.removeTxsOnNextBlockArrive, head.tx)
			}
			// clean data
			txState.Destory()
		*/
	}

}
