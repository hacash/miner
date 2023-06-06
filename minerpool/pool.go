package minerpool

import (
	"fmt"
	"github.com/hacash/chain/leveldb"
	"github.com/hacash/chain/mapset"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	interfaces2 "github.com/hacash/miner/interfaces"
	"github.com/hacash/miner/message"
	"sync"
)

type findBlockMsg struct {
	msg     *message.PowMasterMsg
	account *Account
}

type MinerPool struct {
	Conf *MinerPoolConfig

	currentTcpConnectingCount int32 // Current TCP connections

	blockchain interfaces.BlockChain
	txpool     interfaces.TxPool

	storedb *leveldb.DB

	prevRealtimePeriod    *RealtimePeriod
	currentRealtimePeriod *RealtimePeriod

	/////////////////////////////////////

	checkBlockHeightMiningDict  map[uint64]bool
	successFindNewBlockHashOnce fields.Hash

	successFindNewBlockHashs mapset.Set

	successFindBlockCh chan *findBlockMsg
	//settleRealtimePeriodCh     chan *SettlementPeriod

	/////////////////////////////////////

	status *MinerPoolStatus

	periodChange sync.Mutex
}

func (m *MinerPool) Config() interfaces2.PoWConfig {
	return m.Conf
}

func NewMinerPool(cnf *MinerPoolConfig) *MinerPool {

	db, err := leveldb.OpenFile(cnf.Datadir, nil)
	if err != nil {
		fmt.Println("cnf.Datadir: ", cnf.Datadir)
		panic(err)
	}

	pool := &MinerPool{
		Conf:                        cnf,
		currentTcpConnectingCount:   0,
		checkBlockHeightMiningDict:  make(map[uint64]bool),
		successFindNewBlockHashOnce: nil,
		successFindNewBlockHashs:    mapset.NewSet(),
		successFindBlockCh:          make(chan *findBlockMsg, 4),
		//settleRealtimePeriodCh:     make(chan *SettlementPeriod, 4),
		storedb: db,
		txpool:  nil,
	}

	// read status
	pool.status = pool.readStatus()

	return pool

}

func (p *MinerPool) Start() error {
	if p.blockchain == nil {
		err := fmt.Errorf("p.blockchain not be set yet.")
		return err
	}

	err := p.startServerListen()
	if err != nil {
		return err
	}

	go p.loop()

	return nil
}

func (p *MinerPool) SetBlockChain(blockchain interfaces.BlockChain) {
	if p.blockchain != nil {
		panic("p.blockchain already be set.")
	}
	p.blockchain = blockchain
}

func (p *MinerPool) SetTxPool(tp interfaces.TxPool) {
	p.txpool = tp
}

func (p *MinerPool) GetCurrentAddressCount() int {
	if p.currentRealtimePeriod == nil {
		return 0
	}
	return len(p.currentRealtimePeriod.realtimeAccounts)
}

func (p *MinerPool) GetCurrentMiningAccounts() map[string]*Account {
	if p.currentRealtimePeriod == nil {
		return map[string]*Account{}
	}
	return p.currentRealtimePeriod.realtimeAccounts
}

func (p *MinerPool) GetCurrentTcpConnectingCount() int {
	return int(p.currentTcpConnectingCount)
}

func (p *MinerPool) GetCurrentRealtimePeriod() *RealtimePeriod {
	return p.currentRealtimePeriod
}

func (m *MinerPool) Init() error {
	return nil
}

func (m *MinerPool) DoMining(input interfaces.Block, resCh chan interfaces.Block) error {
	m.Excavate(input, resCh)
	return nil
}
