package diamondminer

import (
	"bytes"
	"fmt"
	"github.com/hacash/core/transactions"
)

// 钻石挖掘自动竞价
func (d *DiamondMiner) doAutoBidForMyDiamond() {
	//fmt.Println("- doAutoBidForMyDiamond")

	firstFeeTxs := d.txpool.GetDiamondCreateTxs(1) // 取出第一枚钻石挖掘交易
	if firstFeeTxs == nil || len(firstFeeTxs) == 0 {
		return // 没有钻石
	}
	firstFeeTx := firstFeeTxs[0]
	// 我自己排第一位
	if bytes.Compare(firstFeeTx.GetAddress(), d.Config.FeeAccount.Address) == 0 {
		if !d.Config.Continued {
			// 非连续挖矿时，停止本机的挖掘
			d.StopAll()
		}
		return
	}
	if d.currentSuccessMiningDiamondTx == nil {
		return
	}
	// 比较钻石序号
	curact := transactions.CheckoutAction_4_DiamondCreateFromTx(d.currentSuccessMiningDiamondTx)
	firstact := transactions.CheckoutAction_4_DiamondCreateFromTx(firstFeeTx)
	if curact == nil || firstact == nil {
		return
	}
	if curact.Number != firstact.Number {
		d.currentSuccessMiningDiamondTx = nil // 无效的挖掘
		return
	}
	// 放弃竞争的地址
	for _, iaddr := range d.Config.AutoBidIgnoreAddresses {
		if bytes.Compare(firstFeeTx.GetAddress(), *iaddr) == 0 {
			return
		}
	}

	// 开始竞价
	topfee := firstFeeTx.GetFee()
	myfee, e1 := topfee.Add(d.Config.AutoBidMarginFee)
	if e1 != nil {
		fmt.Println("doAutoBidForMyDiamond Error:", e1)
		return
	}
	if newmyfee, _, e2 := myfee.CompressForMainNumLen(4, true); e2 == nil && newmyfee != nil {
		myfee = newmyfee // 向上压缩长度
	}
	// 是否高于我设定的最高价
	if d.Config.AutoBidMaxFee.LessThan(&topfee) {
		return
	}
	if d.Config.AutoBidMaxFee.LessThan(myfee) {
		myfee = d.Config.AutoBidMaxFee // 已达到最高价
	}

	// 更新交易费用
	newtx := d.currentSuccessMiningDiamondTx
	newtx.SetFee(myfee)
	// 私钥
	allPrivateKeyBytes := make(map[string][]byte, 1)
	allPrivateKeyBytes[string(d.Config.FeeAccount.Address)] = d.Config.FeeAccount.PrivateKey
	// do sign
	newtx.FillNeedSigns(allPrivateKeyBytes, nil)
	// add to pool
	err4 := d.txpool.AddTx(newtx)
	if err4 != nil {
		fmt.Println("doAutoBidForMyDiamond Add to Tx Pool, Error: ", err4.Error())
		return
	}

	// 成功
	fmt.Printf("\ndiamond auto bid name: <%s>, tx: <%s>, fee: %s => %s .\n\n",
		string(curact.Diamond), newtx.Hash().ToHex(),
		topfee.ToFinString(), myfee.ToFinString(),
	)
}
