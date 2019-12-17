package miner

import (
	"github.com/hacash/core/interfaces"
	"sync/atomic"
)

type Miner struct {
	config *MinerConfig

	blockchain interfaces.BlockChain

	txpool interfaces.TxPool

	powmaster interfaces.PowMaster

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

func (m *Miner) Start() {
	go m.loop()
}

func (m *Miner) StartMining() {
	if m.powmaster == nil {
		panic("[Miner] powmaster is not be set.")
	}
	if atomic.CompareAndSwapUint32(m.isMiningStatus, 0, 1) {
		go m.doStartMining()
	}
}

func (m *Miner) StopMining() {
	if atomic.CompareAndSwapUint32(m.isMiningStatus, 1, 0) {
		go m.doStopMining()
		m.powmaster.StopMining()
	}
}

func (m *Miner) SetBlockChain(bc interfaces.BlockChain) {
	m.blockchain = bc
	bc.SubscribeValidatedBlockOnInsert(m.newBlockOnInsertCh)
}

func (m *Miner) SetPowMaster(pm interfaces.PowMaster) {
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
		err := m.txpool.AddTx(tx)
		if err != nil {

		}
	}()
}
