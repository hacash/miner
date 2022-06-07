package memtxpool

import (
	"github.com/hacash/core/interfaces"
)

func (p *MemTxPool) SubscribeOnAddTxSuccess(addtxCh chan interfaces.Transaction) {
	p.addTxSuccess.Subscribe(addtxCh)
}

// Pause event subscription
func (p *MemTxPool) PauseEventSubscribe() {
	p.changeLock.Lock()
	defer p.changeLock.Unlock()

	p.isBanEventSubscribe = true
}

// Reopen event subscription
func (p *MemTxPool) RenewalEventSubscribe() {
	p.changeLock.Lock()
	defer p.changeLock.Unlock()

	p.isBanEventSubscribe = false
}
