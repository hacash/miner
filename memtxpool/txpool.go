package memtxpool

import (
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/stores"
	"sync"
)

type MemTxPool struct {
	blockchain interfaces.BlockChain

	diamondCreateTxGroup *TxGroup
	simpleTxGroup        *TxGroup

	newDiamondCreateCh chan *stores.DiamondSmelt

	changeLock sync.Mutex

	////////////////////////////////

	txTotalCount uint64
	txTotalSize  uint64

	maxcount uint64
	maxsize  uint64
}

func NewMemTxPool(maxcount, maxsize uint64) *MemTxPool {

	pool := &MemTxPool{
		diamondCreateTxGroup: NewTxGroup(),
		simpleTxGroup:        NewTxGroup(),
		newDiamondCreateCh:   make(chan *stores.DiamondSmelt, 4),
		txTotalCount:         0,
		txTotalSize:          0,
		maxcount:             maxcount,
		maxsize:              maxsize,
	}

	return pool
}

func (p *MemTxPool) Start() {
	go p.loop()
}

func (p *MemTxPool) SetBlockChain(bc interfaces.BlockChain) {

	p.blockchain = bc

	// diamond create event handler
	bc.SubscribeDiamondOnCreate(p.newDiamondCreateCh)

}
