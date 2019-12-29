package memtxpool

import (
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/stores"
	"github.com/hacash/mint/event"
	"sync"
)

type MemTxPool struct {
	blockchain interfaces.BlockChain

	diamondCreateTxGroup *TxGroup
	simpleTxGroup        *TxGroup

	removeTxsOnNextBlockArrive []interfaces.Transaction

	newDiamondCreateCh chan *stores.DiamondSmelt
	newBlockOnInsertCh chan interfaces.Block

	changeLock sync.Mutex

	////////////////////////////////

	automaticallyCleanInvalidTransactions bool

	isBanEventSubscribe bool
	addTxSuccess        event.Feed

	////////////////////////////////

	txTotalCount uint64
	txTotalSize  uint64

	maxcount uint64
	maxsize  uint64
}

func NewMemTxPool(maxcount, maxsize uint64) *MemTxPool {

	pool := &MemTxPool{
		diamondCreateTxGroup:                  NewTxGroup(),
		simpleTxGroup:                         NewTxGroup(),
		newDiamondCreateCh:                    make(chan *stores.DiamondSmelt, 4),
		newBlockOnInsertCh:                    make(chan interfaces.Block, 4),
		txTotalCount:                          0,
		txTotalSize:                           0,
		maxcount:                              maxcount,
		maxsize:                               maxsize,
		isBanEventSubscribe:                   false,
		automaticallyCleanInvalidTransactions: false,
		removeTxsOnNextBlockArrive:            make([]interfaces.Transaction, 0),
	}

	return pool
}

func (p *MemTxPool) Start() {
	go p.loop()
}

func (p *MemTxPool) GetTotalCount() (uint64, uint64) {
	return p.txTotalCount, p.txTotalSize
}

func (p *MemTxPool) GetSimpleTxGroup() *TxGroup {
	return p.simpleTxGroup
}

func (p *MemTxPool) GetDiamondCreateTxGroup() *TxGroup {
	return p.diamondCreateTxGroup
}

func (p *MemTxPool) SetBlockChain(bc interfaces.BlockChain) {

	p.blockchain = bc

	// diamond create event handler
	bc.SubscribeDiamondOnCreate(p.newDiamondCreateCh)
	bc.SubscribeValidatedBlockOnInsert(p.newBlockOnInsertCh)

}

func (p *MemTxPool) GetDiamondCreateTxs() []interfaces.Transaction {
	restxs := make([]interfaces.Transaction, 0)
	if p.diamondCreateTxGroup.Count <= 0 {
		return restxs
	}
	head := p.diamondCreateTxGroup.Head
	for {
		restxs = append(restxs, head.tx)
		head = head.next
		if head == nil {
			break
		}
	}
	return restxs
}

func (p *MemTxPool) SetAutomaticallyCleanInvalidTransactions(set bool) {
	p.automaticallyCleanInvalidTransactions = set
}
