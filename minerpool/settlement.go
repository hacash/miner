package minerpool

import (
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/mint/coinbase"
	"math/big"
	"time"
)

type SettlementPeriod struct {
	period               *RealtimePeriod
	miningSuccessAccount *Account
	successBlockHeight   uint64
	successBlockHash     fields.Hash
}

func (p *MinerPool) createSettlementPeriod(account *Account, period *RealtimePeriod, successBlock interfaces.Block) {
	p.periodChange.Lock()
	defer p.periodChange.Unlock()

	if p.currentRealtimePeriod == period {
		p.prevRealtimePeriod = p.currentRealtimePeriod
		p.currentRealtimePeriod = NewRealtimePeriod(p, nil)
	}

	// add success block hash
	p.successFindNewBlockHashs.Add(string(successBlock.Hash()))

	go func() {
		time.Sleep(time.Second * time.Duration(33))
		p.settleOneSuccessPeriod(&SettlementPeriod{
			period:               period,
			miningSuccessAccount: account,
			successBlockHeight:   successBlock.GetHeight(),
			successBlockHash:     successBlock.Hash(),
		})
	}()

}

func (p *MinerPool) settleOneSuccessPeriod(period *SettlementPeriod) {
	blockHeight := period.successBlockHeight
	// read block from store
	storeBlockHash, err := p.blockchain.GetChainEngineKernel().StateRead().BlockStoreRead().ReadBlockHashByHeight(blockHeight)
	if err != nil {
		return
	}
	if storeBlockHash.Equal(period.successBlockHash) == false {
		return
	}
	// is ok
	totalPowWorth := big.NewInt(0)
	addressPowWorth := make(map[string]*big.Int)
	var minerAccount *Account = nil
	var divrwdAccounts = make([]*Account, 0)
	//fmt.Println("settleOneSuccessPeriod")

	for key, acc := range period.period.realtimeAccounts {
		clients := acc.activeClients.ToSlice()
		for _, cli := range clients {
			cli.(*Client).conn.Close() // Close connection
		}
		//fmt.Println(acc.miningSuccessBlock.Hash().ToHex(), period.successBlockHash.ToHex())
		if acc.miningSuccessBlock != nil && acc.miningSuccessBlock.Hash().Equal(period.successBlockHash) {
			minerAccount = acc // Users who successfully dig out blocks
		}
		// Other miners count the calculation force and copy the value to avoid being modified in the calculation process
		//fmt.Println(acc.address.ToReadable(), acc.realtimePowWorth.String())
		worth := new(big.Int).Add(big.NewInt(0), acc.realtimePowWorth)
		divrwdAccounts = append(divrwdAccounts, acc)
		addressPowWorth[key] = worth
		totalPowWorth = new(big.Int).Add(totalPowWorth, worth)
	}
	if minerAccount == nil {
		//fmt.Println("minerAccount == nil return")
		return
	}
	// Calculate income
	pernum := big.NewInt(10000 * 10000)
	rwdcoin := coinbase.BlockCoinBaseRewardNumber(blockHeight)
	totalReward := int64(rwdcoin) * 10000 * 10000 // 单位：铢
	totalReward = totalReward * int64((1-p.Conf.FeePercentage)*10000) / 10000
	part1of3Reward := totalReward / 3
	part2of3Reward := part1of3Reward * 2
	var rwdAccounts = make([]*Account, 0)
	for _, acc := range divrwdAccounts {
		if totalPowWorth.Cmp(big.NewInt(0)) <= 0 {
			continue
		}
		num1 := new(big.Int).Mul(addressPowWorth[string(acc.address)], pernum)
		num2 := new(big.Int).Div(num1, totalPowWorth)
		reward := num2.Int64() * part2of3Reward / pernum.Int64()
		if acc.address.Equal(minerAccount.address) {
			reward += part1of3Reward // 加上独自的1/3挖出者收益
		}
		//fmt.Println("rwdAccounts = append(rwdAccounts, acc)", acc.address.ToReadable(), reward)
		if reward > 0 {
			// 矿工按比例收益
			rwdAccounts = append(rwdAccounts, acc)
			acc.storeData.appendUnconfirmedRewards(uint32(blockHeight), uint64(reward))
		}
	}
	//fmt.Println(len(divrwdAccounts), len(rwdAccounts), totalPowWorth.String())
	// Save the earnings of the digger
	minerAccount.storeData.findBlocks += 1
	minerAccount.storeData.findCoins += fields.VarUint4(rwdcoin)
	// 保存比例矿工收益
	for _, acc := range rwdAccounts {
		err = p.saveAccountStoreData(acc)
		if err != nil {
			fmt.Println(err)
		}
	}
	// OK settlement completed

	// store success
	_ = p.saveFoundBlockHash(period.successBlockHeight, period.successBlockHash)

}

/*

// Settle one cycle
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
			cli.(*Client).conn.Close() // Close connection
		}
		if acc.miningSuccessBlock != nil {
			minerAccount = acc // Users who successfully dig out blocks
		}
		// Other miners count calculation force and copy values to avoid modification during calculation
		worth := new(big.Int).Add(big.NewInt(0), acc.realtimePowWorth)
		otherAccounts = append(otherAccounts, acc)
		addressPowWorth[key] = worth
		totalPowWorth = new(big.Int).Add(totalPowWorth, worth)

	}
	if minerAccount == nil {
		return
	}
	// Calculate income
	pernum := big.NewInt(10000 * 10000)
	rwdcoin := coinbase.BlockCoinBaseRewardNumber(blockHeight)
	totalReward := int64(rwdcoin) * 10000 * 10000 // 单位：铢
	totalReward = totalReward * int64((1-p.Conf.FeePercentage)*10000) / 10000
	part1of3Reward := totalReward / 3
	part2of3Reward := part1of3Reward * 2
	var rwdAccounts = make([]*Account, 0)
	for _, acc := range otherAccounts {
		if totalPowWorth.Cmp(big.NewInt(0)) == 0 {
			continue
		}
		num1 := new(big.Int).Mul(addressPowWorth[string(acc.address)], pernum)
		num2 := new(big.Int).Div(num1, totalPowWorth)
		reward := num2.Int64() * part2of3Reward / pernum.Int64()
		if reward > 0 {
			rwdAccounts = append(rwdAccounts, acc)
			acc.storeData.appendUnconfirmedRewards(uint32(blockHeight), uint64(reward))
		}
	}
	// Preservation income
	minerAccount.storeData.findBlocks += 1
	minerAccount.storeData.findCoins += fields.VarUint4(rwdcoin)
	if len(rwdAccounts) == 0 {
		// If there is only one account for mining, you will get all rewards
		minerAccount.storeData.appendUnconfirmedRewards(uint32(blockHeight), uint64(totalReward))
	} else {
		minerAccount.storeData.appendUnconfirmedRewards(uint32(blockHeight), uint64(part1of3Reward))
	}
	err = p.saveAccountStoreData(minerAccount)
	for _, acc := range rwdAccounts {
		err = p.saveAccountStoreData(acc)
	}
	// OK settlement completed
	if err != nil {
		fmt.Println(err)
	}

}


*/
