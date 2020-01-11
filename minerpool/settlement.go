package minerpool

import (
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/mint/coinbase"
	"math/big"
)

// 结算一个周期
func (p *MinerPool) settleOnePeriod(period *RealtimePeriod) {
	//p.periodChange.Lock()
	//defer p.periodChange.Unlock()

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
	totalReward = totalReward * int64((1-p.Config.FeePercentage)*10000) / 10000
	partReward := totalReward / 2
	var rwdAccounts = make([]*Account, 0)
	for _, acc := range otherAccounts {
		if totalPowWorth.Cmp(big.NewInt(0)) == 0 {
			continue
		}
		num1 := new(big.Int).Mul(addressPowWorth[string(acc.address)], pernum)
		num2 := new(big.Int).Div(num1, totalPowWorth)
		reward := num2.Int64() * partReward / pernum.Int64()
		if reward > 0 {
			rwdAccounts = append(rwdAccounts, acc)
			acc.storeData.appendUnconfirmedRewards(uint32(blockHeight), uint64(reward))
		}
	}
	// 保存收益
	minerAccount.storeData.findBlocks += 1
	minerAccount.storeData.findCoins += fields.VarInt4(rwdcoin)
	if len(rwdAccounts) == 0 {
		// 如果只有一个账户挖矿，则拿到全部奖励
		minerAccount.storeData.appendUnconfirmedRewards(uint32(blockHeight), uint64(totalReward))
	} else {
		minerAccount.storeData.appendUnconfirmedRewards(uint32(blockHeight), uint64(partReward))
	}
	err = p.saveAccountStoreData(minerAccount)
	for _, acc := range rwdAccounts {
		err = p.saveAccountStoreData(acc)
	}
	// ok 结算完成
	if err != nil {
		fmt.Println(err)
	}

}
