package memtxpool

import (
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
)

func (p *MemTxPool) CheckTxExist(tx interfaces.Transaction) (interfaces.Transaction, bool) {
	return p.CheckTxExistByHash(tx.Hash())
}

func (p *MemTxPool) CheckTxExistByHash(txhash fields.Hash) (interfaces.Transaction, bool) {

	if tx, ok := p.diamondCreateTxGroup.GetItem(string(txhash)); ok {
		return tx.tx, true
	}
	if tx, ok := p.simpleTxGroup.GetItem(string(txhash)); ok {
		return tx.tx, true
	}
	return nil, false

}

func (p *MemTxPool) CopyTxsOrderByFeePurity(targetblockheight uint64, maxcount uint32, maxsize uint32) []interfaces.Transaction {
	p.changeLock.RLock()
	defer p.changeLock.RUnlock()

	restrs := make([]interfaces.Transaction, 0)

	totalcount := uint32(0)
	totalsize := uint32(0)
	var curitxitem *TxItem = nil
	if targetblockheight > 0 && targetblockheight%5 == 0 && p.diamondCreateTxGroup.Count > 0 {
		// pick up all max 100 diamond create tx
		head := p.diamondCreateTxGroup.Head
		for i := 0; i < 100; i++ {
			if head == nil {
				break
			}
			totalcount += 1
			totalsize += head.size
			restrs = append(restrs, head.tx)
			head = head.next
		}
	}
	if p.simpleTxGroup.Head == nil {
		return restrs // simple empty
	}
	curitxitem = p.simpleTxGroup.Head
	for {
		if curitxitem == nil {
			break // end of tail
		}
		totalcount += 1
		totalsize += curitxitem.size
		// check max
		if maxcount > 0 && totalcount > maxcount {
			break
		}
		if maxsize > 0 && totalsize > maxsize {
			break
		}
		// append
		restrs = append(restrs, curitxitem.tx)
		// next
		curitxitem = curitxitem.next
	}

	// results
	return restrs
}
