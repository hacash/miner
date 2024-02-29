package diamondminer

import (
	"bytes"
	"fmt"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/transactions"
)

func (d *DiamondMiner) maybeDiamondMinerStop() {
	if !d.Config.Continued {
		// In case of discontinuous mining, stop the mining of this machine
		//fmt.Println("diamond miner stop all, because fee addr:", iaddr.ToReadable())
		d.StopAll()
	}
}

// Automatic bidding for diamond mining
var diamondalreadypackedtiphei = uint64(0)

func (d *DiamondMiner) doAutoBidForMyDiamond() {
	//fmt.Println("- doAutoBidForMyDiamond")

	if d.currentSuccessMiningDiamondTx == nil {
		return // no mining result yet
	}

	// height check
	curpdblkhei := d.blockchain.GetChainEngineKernel().StateRead().GetPendingBlockHeight()
	if curpdblkhei%5 == 0 {
		// There is no need to bid after the diamond has already been packed
		if diamondalreadypackedtiphei != curpdblkhei {
			fmt.Printf("[diamond auto bid tip]: no need to bid after the diamond has already been packed in block height %d\n",
				curpdblkhei)
		}
		diamondalreadypackedtiphei = curpdblkhei
		return
	}

	firstFeeTxs := d.txpool.GetDiamondCreateTxs(1) // get first bid
	if firstFeeTxs == nil || len(firstFeeTxs) == 0 {
		return // No diamonds
	}
	firstFeeTx := firstFeeTxs[0]
	// Address to give up competition
	for _, iaddr := range d.Config.AutoBidIgnoreAddresses {
		if bytes.Compare(firstFeeTx.GetAddress(), *iaddr) == 0 {
			d.maybeDiamondMinerStop()
			return
		}
	}

	// I came first
	if bytes.Compare(firstFeeTx.GetAddress(), d.Config.FeeAccount.Address) == 0 {
		d.maybeDiamondMinerStop()
		return
	}
	// Compare diamond serial numbers
	curact := transactions.CheckoutAction_4_DiamondCreateFromTx(d.currentSuccessMiningDiamondTx)
	firstact := transactions.CheckoutAction_4_DiamondCreateFromTx(firstFeeTx)
	if curact == nil || firstact == nil {
		return
	}
	if curact.Number != firstact.Number {
		d.currentSuccessMiningDiamondTx = nil // Invalid mining
		return
	}

	// Start bidding
	topfee := firstFeeTx.GetFee()
	myfee, e1 := topfee.Add(d.Config.AutoBidMarginFee)
	if e1 != nil {
		fmt.Println("doAutoBidForMyDiamond Error:", e1)
		return
	}
	if newmyfee, _, e2 := myfee.CompressForMainNumLen(4, true); e2 == nil && newmyfee != nil {
		myfee = newmyfee // Up compression length
	}
	// Is it higher than the maximum price I set
	if d.Config.AutoBidMaxFee.LessThan(topfee) {
		return
	}
	if d.Config.AutoBidMaxFee.LessThan(myfee) {
		myfee = d.Config.AutoBidMaxFee // The highest price has been reached
	}

	// Update transaction fee
	newtx := d.currentSuccessMiningDiamondTx
	newtx.SetFee(myfee)
	newtx.ClearHash() // Reset hash cache
	// Private key
	allPrivateKeyBytes := make(map[string][]byte, 1)
	allPrivateKeyBytes[string(d.Config.FeeAccount.Address)] = d.Config.FeeAccount.PrivateKey
	// do sign
	newtx.FillNeedSigns(allPrivateKeyBytes, nil)
	// add to pool
	err4 := d.txpool.AddTx(newtx.(interfaces.Transaction))
	if err4 != nil {
		fmt.Println("doAutoBidForMyDiamond Add to Tx Pool, Error: ", err4.Error())
		return
	}

	// success do
	fmt.Printf("[diamond auto bid exec]: <%s>, tx: <%s>, fee: %s => %s \n",
		string(curact.Diamond), newtx.Hash().ToHex(),
		topfee.ToFinString(), myfee.ToFinString(),
	)
}
