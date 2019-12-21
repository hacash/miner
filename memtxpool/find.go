package memtxpool

import (
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
)

func (p *MemTxPool) CheckTxExist(tx interfaces.Transaction) bool {
	return p.CheckTxExistByHash(tx.Hash())
}

func (p *MemTxPool) CheckTxExistByHash(txhash fields.Hash) bool {
	if _, ok := p.diamondCreateTxGroup.Items[string(txhash)]; ok {
		return true
	}
	if _, ok := p.simpleTxGroup.Items[string(txhash)]; ok {
		return true
	}
	return false

}

func (p *MemTxPool) CopyTxsOrderByFeePurity(targetblockheight uint64, maxcount uint32, maxsize uint32) []interfaces.Transaction {
	p.changeLock.Lock()
	defer p.changeLock.Unlock()

	restrs := make([]interfaces.Transaction, 0)

	totalcount := uint32(0)
	totalsize := uint32(0)
	var curitxitem *TxItem = nil
	if targetblockheight > 0 && targetblockheight%5 == 0 && p.diamondCreateTxGroup.Count > 0 {
		// pick up one diamond create tx
		curitxitem = p.diamondCreateTxGroup.Head
	}
	if curitxitem == nil {
		if p.simpleTxGroup.Head == nil {
			return restrs // empty
		}
		curitxitem = p.simpleTxGroup.Head
	}
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
