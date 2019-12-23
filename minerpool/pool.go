package minerpool

import (
	"fmt"
	"github.com/hacash/chain/leveldb"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/miner/message"
	"sync"
)

type findBlockMsg struct {
	msg     *message.PowMasterMsg
	account *Account
}

type MinerPool struct {
	Config *MinerPoolConfig

	currentTcpConnectingCount int32 // 当前连接tcp数量

	blockchain interfaces.BlockChain
	txpool     interfaces.TxPool

	storedb *leveldb.DB

	prevRealtimePeriod    *RealtimePeriod
	currentRealtimePeriod *RealtimePeriod

	/////////////////////////////////////

	checkBlockHeightMiningDict    map[uint64]bool
	currentSuccessFindBlockHeight uint64
	successFindBlockCh            chan *findBlockMsg

	/////////////////////////////////////

	status *MinerPoolStatus

	periodChange sync.Mutex
}

func NewMinerPool(cnf *MinerPoolConfig) *MinerPool {

	db, err := leveldb.OpenFile(cnf.Datadir, nil)
	if err != nil {
		fmt.Println("cnf.Datadir: ", cnf.Datadir)
		panic(err)
	}

	pool := &MinerPool{
		Config:                        cnf,
		currentTcpConnectingCount:     0,
		checkBlockHeightMiningDict:    make(map[uint64]bool),
		currentSuccessFindBlockHeight: 0,
		successFindBlockCh:            make(chan *findBlockMsg, 4),
		storedb:                       db,
		txpool:                        nil,
	}

	// read status
	pool.status = pool.readStatus()

	return pool

}

func (p *MinerPool) Start() {
	if p.blockchain == nil {
		panic("p.blockchain not be set yet.")
	}

	err := p.startServerListen()
	if err != nil {
		panic(err)
	}

	go p.loop()
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

func (p *MinerPool) GetCurrentTcpConnectingCount() int32 {
	return p.currentTcpConnectingCount
}
func (p *MinerPool) GetCurrentRealtimePeriod() *RealtimePeriod {
	return p.currentRealtimePeriod
}
