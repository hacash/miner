package minerpool

import (
	"encoding/binary"
	"github.com/hacash/core/fields"
	"github.com/hacash/mint/coinbase"
	"math/big"
)

// 结算一个周期
func (p *MinerPool) settleOnePeriod(period *RealtimePeriod) {
	p.periodChange.Lock()
	defer p.periodChange.Unlock()

	// TODO: settleOnePeriod

	successBlock := period.miningSuccessBlock
	if successBlock == nil {
		return
	}
	blockHeight := successBlock.GetHeight()
	// read block from store
	storeBlockHash, err := p.blockchain.State().BlockStore().ReadBlockHashByHeight(blockHeight)
	if err != nil {
		return
	}
	if storeBlockHash.Equal(successBlock.Hash()) == false {
		return
	}
	// is ok
	totalPowWorth := big.NewInt(0)
	addressPowWorth := make(map[string]*big.Int)
	var minerAccount *Account = nil
	var otherAccounts = make([]*Account, 0)
	for key, acc := range period.realtimeAccounts {
		clients := acc.activeClients.ToSlice()
		for _, cli := range clients {
			cli.(*Client).conn.Close() // 关闭连接
		}
		if acc.miningSuccessBlock != nil {
			minerAccount = acc // 成功挖出区块的用户
		} else {
			// 其他矿工统计算力, 拷贝值，避免运算过程中修改
			worth := new(big.Int).Add(big.NewInt(0), acc.realtimePowWorth)
			otherAccounts = append(otherAccounts, acc)
			addressPowWorth[key] = worth
			totalPowWorth = new(big.Int).Add(totalPowWorth, worth)
		}
	}
	if minerAccount == nil {
		return
	}
	// 计算收益
	pernum := big.NewInt(10000 * 10000)
	rwdcoin := coinbase.BlockCoinBaseRewardNumber(blockHeight)
	totalReward := int64(rwdcoin) * 10000 * 10000 // 单位：铢
	totalReward = totalReward * int64((1-p.config.FeePercentage)*10000) / 10000
	partReward := totalReward / 2
	var rwdAccounts = make([]*Account, 0)
	for _, acc := range otherAccounts {
		num1 := new(big.Int).Mul(addressPowWorth[string(acc.address)], pernum)
		num2 := new(big.Int).Div(num1, totalPowWorth)
		reward := num2.Int64() * partReward
		if reward > 0 {
			rwdAccounts = append(rwdAccounts, acc)
			acc.storeData.UnconfirmedRewards += fields.VarInt8(reward)
			acc.storeData.UnconfirmedRewardListCount += 1
			rwdlstdts := make([]byte, 12)
			binary.BigEndian.PutUint32(rwdlstdts[0:4], uint32(blockHeight))
			binary.BigEndian.PutUint64(rwdlstdts[4:12], uint64(reward))
			acc.storeData.UnconfirmedRewardList = append(acc.storeData.UnconfirmedRewardList, rwdlstdts)
		}
	}
	// 保存收益
	minerAccount.storeData.FindBlocks += 1
	minerAccount.storeData.FindCoins += fields.VarInt4(rwdcoin)
	minerAccount.storeData.UnconfirmedRewards += fields.VarInt8(partReward)
	rwdlstdts := make([]byte, 12)
	binary.BigEndian.PutUint32(rwdlstdts[0:4], uint32(blockHeight))
	binary.BigEndian.PutUint64(rwdlstdts[4:12], uint64(partReward))
	minerAccount.storeData.UnconfirmedRewardListCount += 1
	minerAccount.storeData.UnconfirmedRewardList = append(minerAccount.storeData.UnconfirmedRewardList, rwdlstdts)
	p.saveAccountStoreData(minerAccount)
	for _, acc := range rwdAccounts {
		p.saveAccountStoreData(acc)
	}
	// ok 结算完成

}
