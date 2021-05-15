package diamondminer

import (
	"fmt"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/transactions"
)

func (d *DiamondMiner) successFindDiamondAddTxPool(diamondCreateAction *actions.Action_4_DiamondCreate) {

	tx, e := transactions.NewEmptyTransaction_2_Simple(d.Config.FeeAccount.Address)
	if e != nil {
		return
	}
	tx.Fee = *(d.Config.FeeAmount.Copy())
	//rand.Read(tx.Fee.Numeral)
	//fmt.Println(diamondCreateAction)
	tx.AppendAction(diamondCreateAction)
	// fill sign
	signprivkey := make(map[string][]byte, 0)
	signprivkey[string(d.Config.FeeAccount.Address)] = d.Config.FeeAccount.PrivateKey
	err := tx.FillNeedSigns(signprivkey, nil)
	if err != nil {
		return // error end
	}
	// put into the txpool
	if d.txpool == nil {
		fmt.Println("[Diamond Miner Error] txpool not set")
		return
	}
	err = d.txpool.AddTx(tx)
	if err != nil {
		fmt.Println("[Diamond Miner Error] AddTx error", err)
		return
	}
	d.currentSuccessMiningDiamondTx = tx
	// ok
	fmt.Printf("[Diamond Miner] Diamond %d <%s> add to txpool, hx %s.\n", diamondCreateAction.Number, diamondCreateAction.Diamond, tx.Hash().ToHex())
}
