package memtxpool

import (
	"fmt"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/interfaces"
)

func (p *MemTxPool) AddTx(tx interfaces.Transaction) error {
	p.changeLock.Lock()
	defer p.changeLock.Unlock()

	txitem := &TxItem{
		tx:        tx,
		hash:      tx.Hash(),
		size:      tx.Size(),
		feepurity: tx.FeePurity(),
		diamond:   nil,
	}

	if p.blockchain == nil {
		return fmt.Errorf("[MemTxPool] blockchain is not be set.")
	}

	// check pool max
	if p.maxcount > 0 && p.txTotalCount+1 > p.maxcount {
		return fmt.Errorf("Tx pool max count %d and too mach.", p.maxcount)
	}
	if p.maxsize > 0 && p.txTotalSize+uint64(txitem.size) > p.maxsize {
		return fmt.Errorf("Tx pool max size %d and overflow size.", p.maxsize)
	}

	// check exist
	if havitem := p.diamondCreateTxGroup.Find(txitem.hash); havitem != nil {
		//fmt.Println(havitem.feepurity, txitem.feepurity)
		if txitem.feepurity <= havitem.feepurity {
			return fmt.Errorf("already exist tx %s and fee purity more than or equal the new one.", txitem.hash.ToHex())
		}
		p.diamondCreateTxGroup.RemoveItem(havitem)
	}
	if havitem := p.simpleTxGroup.Find(txitem.hash); havitem != nil {
		//fmt.Println(havitem.feepurity, txitem.feepurity)
		if txitem.feepurity <= havitem.feepurity {
			return fmt.Errorf("already exist tx %s and fee purity more than or equal the new one.", txitem.hash.ToHex())
		}
		if p.simpleTxGroup.RemoveItem(havitem) {
			// sub count
			p.txTotalCount -= 1
			p.txTotalSize -= uint64(havitem.size)
		}
	}
	// do add is diamond ?
	for _, act := range tx.GetActions() {
		if dcact, ok := act.(*actions.Action_4_DiamondCreate); ok {
			err := p.checkDiamondCreate(dcact)
			if err != nil {
				return err
			}
			txitem.diamond = dcact // diamond mark
			p.diamondCreateTxGroup.Add(txitem)
			// feed send
			if p.isBanEventSubscribe == false {
				p.addTxSuccess.Send(tx)
			}
			return nil // add successfully !
		}
	}
	// check tx
	txerr := p.blockchain.ValidateTransaction(tx)
	if txerr != nil {
		return txerr
	}
	// do add simple
	p.simpleTxGroup.Add(txitem)
	// add count
	p.txTotalCount += 1
	p.txTotalSize += uint64(txitem.size)

	// feed send
	if p.isBanEventSubscribe == false {
		p.addTxSuccess.Send(tx)
	}

	fmt.Println("memtxpool add tx successfully:", tx.Hash().ToHex())

	return nil // add successfully !
}
