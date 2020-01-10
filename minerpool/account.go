package minerpool

import (
	"bytes"
	"encoding/binary"
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

func NewEmptyAccountStoreData(curhei uint64) *AccountStoreData {
	return &AccountStoreData{
		0,
		0,
		0,
		0,
		0,
		fields.VarInt4(uint32(curhei)),
		0,
		[]fields.Bytes12{},
		fields.Bytes16{},
	}
}

func (s *AccountStoreData) moveRewards(target string, rewards uint64) bool {
	if target == "deserved" {
		if uint64(s.unconfirmedRewards) < rewards {
			return false
		}
		s.unconfirmedRewards = fields.VarInt8(uint64(s.unconfirmedRewards) - rewards)
		s.deservedRewards = fields.VarInt8(uint64(s.deservedRewards) + rewards)
		return true
	} else if target == "complete" {
		if uint64(s.deservedRewards) < rewards {
			return false
		}
		s.deservedRewards = fields.VarInt8(uint64(s.deservedRewards) - rewards)
		s.completeRewards = fields.VarInt8(uint64(s.completeRewards) + rewards)
		return true
	}
	return false
}

func (s *AccountStoreData) unshiftUnconfirmedRewards(lessthanblkhei uint64) (uint32, uint64, bool) {
	if s.unconfirmedRewardListCount == 0 {
		return 0, 0, false
	}
	valdts := s.unconfirmedRewardList[0]
	blockhei := binary.BigEndian.Uint32(valdts[0:4])
	if lessthanblkhei > 0 && uint64(blockhei) > lessthanblkhei {
		return 0, 0, false
	}
	s.unconfirmedRewardListCount -= 1
	s.unconfirmedRewardList = s.unconfirmedRewardList[1:]
	// ok return
	return blockhei,
		binary.BigEndian.Uint64(valdts[4:12]),
		true
}

func (s *AccountStoreData) appendUnconfirmedRewards(blockHeight uint32, rewards uint64) {
	s.unconfirmedRewards += fields.VarInt8(rewards)
	s.unconfirmedRewardListCount += 1
	rwdlstdts := make([]byte, 12)
	binary.BigEndian.PutUint32(rwdlstdts[0:4], uint32(blockHeight))
	binary.BigEndian.PutUint64(rwdlstdts[4:12], uint64(rewards))
	s.unconfirmedRewardList = append(s.unconfirmedRewardList, rwdlstdts)
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
	b8, _ := s.unconfirmedRewardListCount.Serialize()
	buf.Write(b8)
	for i := 0; i < int(s.unconfirmedRewardListCount); i++ {
		b, _ := s.unconfirmedRewardList[i].Serialize()
		buf.Write(b)
	}
	///
	b9, _ := s.others.Serialize()
	buf.Write(b9)
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
		curblkhei := uint64(0)
		if p.currentRealtimePeriod.targetBlock != nil {
			curblkhei = p.currentRealtimePeriod.targetBlock.GetHeight()
		}
		accstodts := p.loadAccountStoreData(curblkhei, address)
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
