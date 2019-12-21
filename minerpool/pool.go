package minerpool

import (
	"fmt"
	"github.com/hacash/chain/leveldb"
	"github.com/hacash/core/interfaces"
	"sync"
)

type MinerPool struct {
	Config *MinerPoolConfig

	currentTcpConnectingCount int32 // 当前连接tcp数量

	blockchain interfaces.BlockChain

	storedb *leveldb.DB

	prevRealtimePeriod    *RealtimePeriod
	currentRealtimePeriod *RealtimePeriod

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
		Config:                    cnf,
		currentTcpConnectingCount: 0,
		storedb:                   db,
	}

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

func (p *MinerPool) GetCurrentTcpConnectingCount() int32 {
	return p.currentTcpConnectingCount
}
func (p *MinerPool) GetCurrentRealtimePeriod() *RealtimePeriod {
	return p.currentRealtimePeriod
}
