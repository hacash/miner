package miner

import (
	"fmt"
	"github.com/hacash/core/interfaces"
	itfcs "github.com/hacash/miner/interfaces"
	"sync/atomic"
)

type Miner struct {
	config *MinerConfig

	blockchain interfaces.BlockChain

	txpool interfaces.TxPool

	powmaster itfcs.PoWMaster

	isMiningStatus     *uint32
	stopSignCh         chan bool
	newBlockOnInsertCh chan interfaces.Block
}

func NewMiner(cnf *MinerConfig) *Miner {
	miner := &Miner{
		config:             cnf,
		stopSignCh:         make(chan bool, 1),
		newBlockOnInsertCh: make(chan interfaces.Block, 4),
	}
	var sm uint32 = 0
	miner.isMiningStatus = &sm
	return miner
}

func (m *Miner) Start() error {
	go m.loop()
	return nil
}

func (m *Miner) StartMining() error {
	if m.powmaster == nil {
		return fmt.Errorf("[Miner] powmaster is not be set.")
	}
	if atomic.CompareAndSwapUint32(m.isMiningStatus, 0, 1) {
		go m.doStartMining()
	}
	return nil
}

func (m *Miner) StopMining() {
	if atomic.CompareAndSwapUint32(m.isMiningStatus, 1, 0) {
		go m.doStopMining()
		m.powmaster.StopMining()
	}
}

func (m *Miner) SetBlockChain(bc interfaces.BlockChain) {
	m.blockchain = bc
	bc.GetChainEngineKernel().SubscribeValidatedBlockOnInsert(m.newBlockOnInsertCh)
}

func (m *Miner) SetPowServer(pm itfcs.PoWMaster) {
	m.powmaster = pm
}

func (m *Miner) SetTxPool(tp interfaces.TxPool) {
	m.txpool = tp
}

func (m *Miner) SubmitTx(tx interfaces.Transaction) {
	if m.blockchain == nil {
		panic("[Miner] blockchain is not be set.")
	}
	if m.txpool == nil {
		panic("[Miner] txpool is not be set.")
	}
	// add tx to pool
	go func() {
		err := m.txpool.AddTx(tx.(interfaces.Transaction))
		if err != nil {

		}
	}()
}
