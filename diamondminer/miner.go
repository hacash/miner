package diamondminer

import (
	"fmt"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/genesis"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/stores"
	"os"
	"sync"
	"time"
)

type DiamondMiner struct {
	Config *DiamondMinerConfig

	blockchain interfaces.BlockChain
	txpool     interfaces.TxPool

	stopMarks map[*byte]*byte

	newDiamondBeFoundCh    chan *stores.DiamondSmelt
	successMiningDiamondCh chan *actions.Action_4_DiamondCreate

	// 当前挖掘成功的钻石交易
	currentSuccessMiningDiamondTx interfaces.Transaction

	changeLock sync.Mutex
}

func NewDiamondMiner(cnf *DiamondMinerConfig) *DiamondMiner {
	dia := &DiamondMiner{
		Config:                        cnf,
		stopMarks:                     map[*byte]*byte{},
		newDiamondBeFoundCh:           make(chan *stores.DiamondSmelt, 2),
		successMiningDiamondCh:        make(chan *actions.Action_4_DiamondCreate, 4),
		currentSuccessMiningDiamondTx: nil,
	}

	return dia
}

func (d *DiamondMiner) Start() {
	if d.blockchain == nil {
		panic("d.blockchain not be set yet.")
	}
	if d.txpool == nil {
		panic("d.txpool not be set yet.")
	}

	go d.loop()

	go func() {
		time.Sleep(time.Second)
		prev, e := d.blockchain.State().ReadLastestDiamond()
		if e != nil {
			fmt.Println("[Diamond Miner Error] miner cannot start: ", e)
			os.Exit(0)
		}
		// is first
		if prev == nil {
			genesisblk := genesis.GetGenesisBlock()
			prev = &stores.DiamondSmelt{
				Number:           0,
				ContainBlockHash: genesisblk.Hash(),
			}
		}
		// do mining
		d.RunMining(prev, d.successMiningDiamondCh)
	}()
}

func (m *DiamondMiner) SetTxPool(tp interfaces.TxPool) {
	m.txpool = tp
}

func (d *DiamondMiner) SetBlockChain(blockchain interfaces.BlockChain) {
	if d.blockchain != nil {
		panic("d.blockchain already be set.")
	}
	d.blockchain = blockchain
	// feed event
	blockchain.SubscribeDiamondOnCreate(d.newDiamondBeFoundCh)
}
