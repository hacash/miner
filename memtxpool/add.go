package memtxpool

import (
	"fmt"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/interfaces"
	"time"
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

	// check tx time
	if tx.GetTimestamp() > uint64(time.Now().Unix()) {
		return fmt.Errorf("tx timestamp cannot more than now.")
	}

	// check pool max
	if p.maxcount > 0 && p.txTotalCount+1 > p.maxcount {
		return fmt.Errorf("Tx pool max count %d and too mach.", p.maxcount)
	}
	if p.maxsize > 0 && p.txTotalSize+uint64(txitem.size) > p.maxsize {
		return fmt.Errorf("Tx pool max size %d and overflow size.", p.maxsize)
	}

	// 是否为全新首次添加
	isTxFirstAdd := true

	// check exist
	if havitem := p.diamondCreateTxGroup.Find(txitem.hash); havitem != nil {
		//fmt.Println(havitem.feepurity, txitem.feepurity)
		if txitem.feepurity <= havitem.feepurity {
			return fmt.Errorf("already exist tx %s and fee purity more than or equal the new one.", txitem.hash.ToHex())
		}
		// check fee
		txfee := txitem.tx.GetFee()
		febls := p.blockchain.State().Balance(txitem.tx.GetAddress())
		blastr := "ㄜ0:0"
		if febls != nil {
			blastr = febls.Amount.ToFinString()
		}
		if febls == nil || febls.Amount.LessThan(&txfee) {
			// 余额不足以支付手续费
			return fmt.Errorf("fee address balance need not less than %s but got %s.", txfee.ToFinString(), txitem.tx.GetAddress(), blastr)
		}
		// check ok
		p.diamondCreateTxGroup.RemoveItem(havitem)
		isTxFirstAdd = false
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
		isTxFirstAdd = false
	}
	// do add is diamond ?
	for _, act := range tx.GetActions() {
		if dcact, ok := act.(*actions.Action_4_DiamondCreate); ok {
			// is diamond create trs
			err := p.checkDiamondCreate(tx, dcact)
			if err != nil {
				return err
			}
			txitem.diamond = dcact // diamond mark
			p.diamondCreateTxGroup.Add(txitem)
			// feed send
			if p.isBanEventSubscribe == false {
				p.addTxSuccess.Send(tx)
			}
			if isTxFirstAdd {
				fmt.Println("memtxpool add diamond create tx:", tx.Hash().ToHex(), ", diamond:", dcact.Number, string(dcact.Diamond))
			}
			return nil // add successfully !
		}
	}
	// check tx
	txerr := p.blockchain.ValidateTransaction(tx, func(tmpState interfaces.ChainState) {
		// 标记是矿池中验证tx
		tmpState.SetInMemTxPool(true)
	})
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

	if isTxFirstAdd {
		fmt.Println("memtxpool add tx:", tx.Hash().ToHex())
	}

	return nil // add successfully !
}
