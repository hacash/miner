package miningpool

import (
	"github.com/hacash/chain/leveldb"
	"github.com/hacash/core/fields"
	"math/big"
)

type MinerPool struct {
	config *MinerPoolConfig

	realtimeAccounts      map[string]*Account // [*Account]
	realtimeTotalPowWorth big.Int             // 周期内算力总统计

	storedb *leveldb.DB

	/////////////////////////////////////

	FindBlocks fields.VarInt4 // 挖出的区块数量
	FindCoins  fields.VarInt4 // 挖出的币数量

}

func NewMinerPool(cnf *MinerPoolConfig) *MinerPool {
	pool := &MinerPool{
		config:                cnf,
		realtimeAccounts:      make(map[string]*Account),
		realtimeTotalPowWorth: big.Int{},
	}

	return pool

}

func (p *MinerPool) Start() {

	go p.loop()
}
