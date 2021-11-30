package miner

import (
	"fmt"
	"github.com/hacash/core/interfacev2"
	"sync/atomic"
)

type Miner struct {
	config *MinerConfig

	blockchain interfacev2.BlockChain

	txpool interfacev2.TxPool

	powserver interfacev2.PowServer

	isMiningStatus     *uint32
	stopSignCh         chan bool
	newBlockOnInsertCh chan interfacev2.Block
}

func NewMiner(cnf *MinerConfig) *Miner {
	miner := &Miner{
		config:             cnf,
		stopSignCh:         make(chan bool, 1),
		newBlockOnInsertCh: make(chan interfacev2.Block, 4),
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
	if m.powserver == nil {
		return fmt.Errorf("[Miner] powserver is not be set.")
	}
	if atomic.CompareAndSwapUint32(m.isMiningStatus, 0, 1) {
		go m.doStartMining()
	}
	return nil
}

func (m *Miner) StopMining() {
	if atomic.CompareAndSwapUint32(m.isMiningStatus, 1, 0) {
		go m.doStopMining()
		m.powserver.StopMining()
	}
}

func (m *Miner) SetBlockChain(bc interfacev2.BlockChain) {
	m.blockchain = bc
	bc.SubscribeValidatedBlockOnInsert(m.newBlockOnInsertCh)
}

func (m *Miner) SetPowServer(pm interfacev2.PowServer) {
	m.powserver = pm
}

func (m *Miner) SetTxPool(tp interfacev2.TxPool) {
	m.txpool = tp
}

func (m *Miner) SubmitTx(tx interfacev2.Transaction) {
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
