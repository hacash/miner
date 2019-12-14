package memtxpool

import (
	"github.com/hacash/core/interfaces"
	"sync"
)

type MemTxPool struct {
	blockchain interfaces.BlockChain

	diamondCreateTxs *TxItem
	Txs              *TxItem

	changeLock sync.Mutex
}
