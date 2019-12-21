package memtxpool

import (
	"github.com/hacash/core/interfaces"
)

func (p *MemTxPool) SubscribeOnAddTxSuccess(addtxCh chan interfaces.Transaction) {
	p.addTxSuccess.Subscribe(addtxCh)
}
