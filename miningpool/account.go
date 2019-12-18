package miningpool

import (
	"bytes"
	"fmt"
	"github.com/hacash/chain/mapset"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	"math/big"
	"sync"
)

type Account struct {
	miningSuccessBlockHash fields.Hash

	realtimePeriod *RealtimePeriod // 所属统计周期

	address fields.Address // 获得奖励地址

	workBlock interfaces.Block

	activeClients mapset.Set // [*Client] // 正在连接的客户端

	realtimePowWorth *big.Int // 周期内算力统计

	///////////////////////////////////////////////

	storeData *AccountStoreData

	change sync.Mutex
}

func NewAccountByPeriod(address fields.Address, period *RealtimePeriod) *Account {
	acc := &Account{
		miningSuccessBlockHash: nil,
		realtimePeriod:         period,
		address:                address,
		workBlock:              period.targetBlock,
		activeClients:          mapset.NewSet(),
		realtimePowWorth:       new(big.Int),
		storeData:              nil,
	}
	return acc
}

func (a *Account) CopyByPeriod(period *RealtimePeriod) *Account {
	acc := &Account{
		realtimePeriod:   period,
		address:          a.address,
		workBlock:        period.targetBlock,
		activeClients:    mapset.NewSet(),
		realtimePowWorth: new(big.Int),
		storeData:        a.storeData,
	}
	return acc
}

const (
	AccountStoreDataSizeRealUsed = 4 + 4 + 8 + 8 + 4
	AccountStoreDataSize         = AccountStoreDataSizeRealUsed + (4 * 16)
)

type AccountStoreData struct {
	FindBlocks              fields.VarInt4 // 挖出的区块数量
	FindCoins               fields.VarInt4 // 挖出的币数量
	CompleteRewards         fields.VarInt8 // 已完成并打币的奖励  单位： ㄜ240  （10^8）
	DeservedRewards         fields.VarInt8 // 应得但还没有打币的奖励  单位： ㄜ240  （10^8）
	PrevTransferBlockHeight fields.VarInt4 // 上一次打币时的区块
}

func NewEmptyAccountStoreData() *AccountStoreData {
	return &AccountStoreData{0, 0, 0, 0, 0}
}

func (s *AccountStoreData) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	b1, _ := s.FindBlocks.Serialize()
	b2, _ := s.FindCoins.Serialize()
	b3, _ := s.CompleteRewards.Serialize()
	b4, _ := s.DeservedRewards.Serialize()
	b5, _ := s.PrevTransferBlockHeight.Serialize()
	buf.Write(b1)
	buf.Write(b2)
	buf.Write(b3)
	buf.Write(b4)
	buf.Write(b5)
	resbuf := make([]byte, AccountStoreDataSize)
	copy(resbuf, buf.Bytes())
	return resbuf, nil
}

func (s *AccountStoreData) Parse(buf []byte, seek uint32) (uint32, error) {
	if uint32(len(buf))-seek < AccountStoreDataSizeRealUsed {
		return 0, fmt.Errorf("size error.")
	}
	seek, _ = s.FindBlocks.Parse(buf, seek)
	seek, _ = s.FindCoins.Parse(buf, seek)
	seek, _ = s.CompleteRewards.Parse(buf, seek)
	seek, _ = s.DeservedRewards.Parse(buf, seek)
	seek, _ = s.PrevTransferBlockHeight.Parse(buf, seek)
	return seek, nil
}

func (s *AccountStoreData) Size() uint32 {
	return AccountStoreDataSize
}

////////////////////////////////////////////////////////////

func (p *MinerPool) loadAccountAndAddPeriodByAddress(address fields.Address) *Account {
	p.periodChange.Lock()
	defer p.periodChange.Unlock()
	// check current
	if p.currentRealtimePeriod != nil {
		for key, acc := range p.currentRealtimePeriod.realtimeAccounts {
			if key == string(address) {
				return acc
			}
		}
	}
	// copy
	if p.prevRealtimePeriod != nil {
		for key, acc := range p.prevRealtimePeriod.realtimeAccounts {
			if key == string(address) {
				newacc := acc.CopyByPeriod(p.currentRealtimePeriod)
				p.currentRealtimePeriod.realtimeAccounts[key] = newacc // copy add
				return newacc
			}
		}
	}
	// create
	if p.currentRealtimePeriod != nil {
		accstodts := p.loadAccountStoreData(address)
		newacc := NewAccountByPeriod(address, p.currentRealtimePeriod)
		newacc.storeData = accstodts
		p.currentRealtimePeriod.realtimeAccounts[string(address)] = newacc
		return newacc
	}
	// not yet
	return nil
}
