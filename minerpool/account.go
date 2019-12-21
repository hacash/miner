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

func (a Account) CopyByPeriod(period *RealtimePeriod) *Account {
	acc := &Account{
		realtimePeriod:   period,
		address:          append([]byte{}, a.address...),
		workBlock:        period.targetBlock,
		activeClients:    mapset.NewSet(),
		realtimePowWorth: new(big.Int),
		storeData:        a.storeData,
	}
	return acc
}

func (a *Account) GetAddress() fields.Address {
	return a.address
}

func (a *Account) GetStoreData() *AccountStoreData {
	return a.storeData
}

func (a *Account) GetClientCount() int {
	return a.activeClients.Cardinality()
}

func (a *Account) GetRealtimePowWorth() *big.Int {
	return big.NewInt(0).Set(a.realtimePowWorth)
}

/////////////////////////////

type AccountStoreData struct {
	//
	findBlocks              fields.VarInt4 // 挖出的区块数量
	findCoins               fields.VarInt4 // 挖出的币数量
	completeRewards         fields.VarInt8 // 已完成并打币的奖励     单位：铢 ㄜ240  （10^8）
	deservedRewards         fields.VarInt8 // 应得但还没有打币的奖励  单位：铢 ㄜ240  （10^8）
	unconfirmedRewards      fields.VarInt8 // 挖出还没经过确认的奖励  单位：铢 ㄜ240  （10^8）
	prevTransferBlockHeight fields.VarInt4 // 上一次打币时的区块
	//
	unconfirmedRewardListCount fields.VarInt4
	unconfirmedRewardList      []fields.Bytes12 // 4 + 8 : blockHeight + reward
	//
	others fields.Bytes16 // 备用扩展字段
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

func (s *AccountStoreData) GetFinds() (int, int) {
	return int(s.findBlocks), int(s.findCoins)
}

func (s *AccountStoreData) GetRewards() (int64, int64, int64) {
	return int64(s.completeRewards), int64(s.deservedRewards), int64(s.unconfirmedRewards)
}

func (s *AccountStoreData) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	b1, _ := s.findBlocks.Serialize()
	buf.Write(b1)
	b2, _ := s.findCoins.Serialize()
	buf.Write(b2)
	b3, _ := s.completeRewards.Serialize()
	buf.Write(b3)
	b4, _ := s.deservedRewards.Serialize()
	buf.Write(b4)
	b5, _ := s.unconfirmedRewards.Serialize()
	buf.Write(b5)
	b6, _ := s.prevTransferBlockHeight.Serialize()
	buf.Write(b6)
	b7, _ := s.unconfirmedRewards.Serialize()
	buf.Write(b7)
	for i := 0; i < int(s.unconfirmedRewardListCount); i++ {
		b, _ := s.unconfirmedRewardList[i].Serialize()
		buf.Write(b)
	}
	///
	b8, _ := s.others.Serialize()
	buf.Write(b8)
	return buf.Bytes(), nil
}

func (s *AccountStoreData) Parse(buf []byte, seek uint32) (uint32, error) {
	seek, _ = s.findBlocks.Parse(buf, seek)
	seek, _ = s.findCoins.Parse(buf, seek)
	seek, _ = s.completeRewards.Parse(buf, seek)
	seek, _ = s.deservedRewards.Parse(buf, seek)
	seek, _ = s.unconfirmedRewards.Parse(buf, seek)
	seek, _ = s.prevTransferBlockHeight.Parse(buf, seek)
	seek, _ = s.unconfirmedRewardListCount.Parse(buf, seek)
	s.unconfirmedRewardList = make([]fields.Bytes12, s.unconfirmedRewardListCount)
	for i := 0; i < int(s.unconfirmedRewardListCount); i++ {
		_, _ = s.unconfirmedRewardList[i].Parse(buf, seek)
		seek += 12
	}
	seek, _ = s.others.Parse(buf, seek)
	return seek, nil
}

func (s *AccountStoreData) Size() uint32 {
	return 4 + 4 + 8 + 8 + 8 + 4 +
		4 + uint32(s.unconfirmedRewardListCount*12) +
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
		//fmt.Println("newacc.storeData = accstodts  ", newacc.address.ToReadable())
		newacc.storeData = accstodts
		//fmt.Println("p.currentRealtimePeriod.realtimeAccounts[string(address)] = newacc")
		p.currentRealtimePeriod.realtimeAccounts[string(address)] = newacc
		return newacc
	}
	// not yet
	return nil
}
