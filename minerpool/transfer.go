package minerpool

import (
	"fmt"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/transactions"
	"math/big"
)

// 开始判断后打币
func (p *MinerPool) startDoTransfer(curblkheight uint64, period *RealtimePeriod) {
	p.periodChange.Lock()
	defer p.periodChange.Unlock()

	if curblkheight%5 != 0 {
		return
	}
	trsAccounts := filterOutCanBeTransferred(curblkheight, period)
	if len(trsAccounts) == 0 {
		return // empty
	}
	// create tx
	tx, err := transactions.NewEmptyTransaction_2_Simple(p.Config.RewardAccount.Address)
	if err != nil {
		return // error
	}
	//transfers := make([]*actions.Action_1_SimpleTransfer, 0)
	baseFee := fields.NewAmountSmall(5, 243) // base fee
	totalFee := baseFee.Copy()
	totalAmount := fields.NewEmptyAmount()
	for _, acc := range trsAccounts {
		amt, _ := fields.NewAmountByBigIntWithUnit(big.NewInt(int64(acc.storeData.deservedRewards)), 240)
		totalAmount, _ = totalAmount.Add(amt)
		totalFee, _ = totalFee.Add(baseFee)
		trsact := actions.NewAction_1_SimpleTransfer(acc.address, amt)
		//transfers = append(transfers, trsact)
		tx.AppendAction(trsact)
	}
	// check balance
	checkAmt, _ := totalAmount.Add(totalFee)
	balance := p.blockchain.State().Balance(p.Config.RewardAccount.Address)
	if balance == nil {
		fmt.Printf("[Miner Pool Transfer Error] Balance not is empty.")
		return
	}
	if balance.Amount.LessThan(checkAmt) {
		fmt.Printf("[Miner Pool Transfer Error] Balance not enough, need %s but only have %s .", checkAmt.ToFinString(), balance.Amount.ToFinString())
		return
	}
	tx.Fee = *totalFee // set fee
	// fill sign
	signprivkey := make(map[string][]byte, 0)
	signprivkey[string(p.Config.RewardAccount.Address)] = p.Config.RewardAccount.PrivateKey
	err = tx.FillNeedSigns(signprivkey, nil)
	if err != nil {
		return // error end
	}
	// put into the txpool
	if p.txpool == nil {
		fmt.Println("[Miner Pool Transfer Error] txpool not set")
		return
	}
	err = p.txpool.AddTx(tx)
	if err != nil {
		fmt.Println("[Miner Pool Transfer Error] AddTx error", err)
		return
	}
	// store
	err = p.saveTransferTx(tx)
	if err != nil {
		fmt.Println("[Miner Pool Transfer Error] saveTransferTx error", err)
		return
	}
	// change store data
	for _, acc := range trsAccounts {
		acc.storeData.moveRewards("complete", uint64(acc.storeData.deservedRewards))
		p.saveAccountStoreData(acc)
	}
	// ok
	fmt.Printf(" --> --> --> --> miner pool transfer to %d address cost amount %s fee %s .\n", len(tx.GetActions()), totalAmount.ToFinString(), totalFee.ToFinString())
}

func filterOutCanBeTransferred(curblkheight uint64, period *RealtimePeriod) []*Account {
	resaccs := make([]*Account, 0)
	for _, acc := range period.realtimeAccounts {
		rwdamt := uint64(acc.storeData.deservedRewards)
		if rwdamt > uint64(2)*10000*10000 {
			resaccs = append(resaccs, acc)
		} else if rwdamt > uint64(2)*10000*1000 {
			if uint64(acc.storeData.prevTransferBlockHeight)+100 < curblkheight {
				resaccs = append(resaccs, acc)
			}
		} else if rwdamt > uint64(2)*10000*100 {
			if uint64(acc.storeData.prevTransferBlockHeight)+1000 < curblkheight {
				resaccs = append(resaccs, acc)
			}
		}
	}

	return resaccs
}
