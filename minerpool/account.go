package minerpool

import (
	"bytes"
	"github.com/hacash/chain/mapset"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	"math/big"
	"sync"
)

type Account struct {
	miningSuccessBlock interfaces.Block

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
		miningSuccessBlock: nil,
		realtimePeriod:     period,
		address:            address,
		workBlock:          period.targetBlock,
		activeClients:      mapset.NewSet(),
		realtimePowWorth:   new(big.Int),
		storeData:          nil,
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

type AccountStoreData struct {
	//
	FindBlocks              fields.VarInt4 // 挖出的区块数量
	FindCoins               fields.VarInt4 // 挖出的币数量
	CompleteRewards         fields.VarInt8 // 已完成并打币的奖励     单位：铢 ㄜ240  （10^8）
	DeservedRewards         fields.VarInt8 // 应得但还没有打币的奖励  单位：铢 ㄜ240  （10^8）
	UnconfirmedRewards      fields.VarInt8 // 挖出还没经过确认的奖励  单位：铢 ㄜ240  （10^8）
	PrevTransferBlockHeight fields.VarInt4 // 上一次打币时的区块
	//
	UnconfirmedRewardListCount fields.VarInt4
	UnconfirmedRewardList      []fields.Bytes12 // 4 + 8 : blockHeight + reward
	//
	Others fields.Bytes16 // 备用扩展字段
}

func NewEmptyAccountStoreData() *AccountStoreData {
	return &AccountStoreData{
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		[]fields.Bytes12{},
		fields.Bytes16{},
	}
}

func (s *AccountStoreData) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	b1, _ := s.FindBlocks.Serialize()
	buf.Write(b1)
	b2, _ := s.FindCoins.Serialize()
	buf.Write(b2)
	b3, _ := s.CompleteRewards.Serialize()
	buf.Write(b3)
	b4, _ := s.DeservedRewards.Serialize()
	buf.Write(b4)
	b5, _ := s.UnconfirmedRewards.Serialize()
	buf.Write(b5)
	b6, _ := s.PrevTransferBlockHeight.Serialize()
	buf.Write(b6)
	b7, _ := s.UnconfirmedRewards.Serialize()
	buf.Write(b7)
	for i := 0; i < int(s.UnconfirmedRewardListCount); i++ {
		b, _ := s.UnconfirmedRewardList[i].Serialize()
		buf.Write(b)
	}
	///
	b8, _ := s.Others.Serialize()
	buf.Write(b8)
	return buf.Bytes(), nil
}

func (s *AccountStoreData) Parse(buf []byte, seek uint32) (uint32, error) {
	seek, _ = s.FindBlocks.Parse(buf, seek)
	seek, _ = s.FindCoins.Parse(buf, seek)
	seek, _ = s.CompleteRewards.Parse(buf, seek)
	seek, _ = s.DeservedRewards.Parse(buf, seek)
	seek, _ = s.UnconfirmedRewards.Parse(buf, seek)
	seek, _ = s.PrevTransferBlockHeight.Parse(buf, seek)
	seek, _ = s.UnconfirmedRewardListCount.Parse(buf, seek)
	s.UnconfirmedRewardList = make([]fields.Bytes12, s.UnconfirmedRewardListCount)
	for i := 0; i < int(s.UnconfirmedRewardListCount); i++ {
		_, _ = s.UnconfirmedRewardList[i].Parse(buf, seek)
		seek += 12
	}
	seek, _ = s.Others.Parse(buf, seek)
	return seek, nil
}

func (s *AccountStoreData) Size() uint32 {
	return 4 + 4 + 8 + 8 + 8 + 4 +
		4 + uint32(s.UnconfirmedRewardListCount*12) +
		16
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
