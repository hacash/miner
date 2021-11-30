package memtxpool

import (
	"github.com/hacash/core/interfacev2"
	"github.com/hacash/core/stores"
	"github.com/hacash/mint/event"
	"sync"
)

type MemTxPool struct {
	blockchain interfacev2.BlockChain

	diamondCreateTxGroup *TxGroup
	simpleTxGroup        *TxGroup

	removeTxsOnNextBlockArrive []interfacev2.Transaction

	newDiamondCreateCh chan *stores.DiamondSmelt
	newBlockOnInsertCh chan interfacev2.Block

	changeLock sync.RWMutex

	////////////////////////////////

	automaticallyCleanInvalidTransactions bool // 是否自动清理失效的交易

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
		newBlockOnInsertCh:                    make(chan interfacev2.Block, 4),
		txTotalCount:                          0,
		txTotalSize:                           0,
		maxcount:                              maxcount,
		maxsize:                               maxsize,
		isBanEventSubscribe:                   false,
		automaticallyCleanInvalidTransactions: false,
		removeTxsOnNextBlockArrive:            make([]interfacev2.Transaction, 0),
		changeLock:                            sync.RWMutex{},
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

func (p *MemTxPool) SetBlockChain(bc interfacev2.BlockChain) {

	p.blockchain = bc

	// diamond create event handler
	bc.SubscribeDiamondOnCreate(p.newDiamondCreateCh)
	bc.SubscribeValidatedBlockOnInsert(p.newBlockOnInsertCh)

}

func (p *MemTxPool) GetDiamondCreateTxs(num int) []interfacev2.Transaction {
	p.changeLock.RLock()
	defer p.changeLock.RUnlock()

	restxs := make([]interfacev2.Transaction, 0)
	if p.diamondCreateTxGroup.Count <= 0 {
		return restxs
	}
	head := p.diamondCreateTxGroup.Head
	for {
		restxs = append(restxs, head.tx)
		if num > 0 {
			if len(restxs) >= num {
				break // 控制数量
			}
		}
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
