package memtxpool

import (
	"fmt"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/mint"
	"time"
)

func (p *MemTxPool) checkDiamondCreate(newtx interfaces.Transaction, act *actions.Action_4_DiamondCreate) error {

	newtxhash := newtx.Hash()
	txhxhex := newtxhash.ToHex()
	blockstate := p.blockchain.GetChainEngineKernel().StateRead()
	exist, e0 := blockstate.CheckTxHash(newtxhash)
	//fmt.Println(exist, exist_tx_bytes)
	if e0 != nil {
		return e0
	}
	if exist {
		return fmt.Errorf("diamond create tx %s is exist in blockchain.", txhxhex)
	}
	// check
	if newtx.GetTimestamp() > uint64(time.Now().Unix()) {
		return fmt.Errorf("diamond create tx %s timestamp cannot more than now.", txhxhex)
	}
	// fee purity
	if newtx.FeePurity() < mint.MinTransactionFeePurityOfOneByte {
		return fmt.Errorf("diamond create tx %s handling fee is too low for miners to accept.", txhxhex)
	}
	// sign
	ok, e1 := newtx.VerifyAllNeedSigns()
	if !ok || e1 != nil {
		return fmt.Errorf("diamond create tx %s verify signature error", txhxhex)
	}
	// 检查余额 // check fee
	txfee := newtx.GetFee()
	febls, e := p.blockchain.GetChainEngineKernel().StateRead().Balance(newtx.GetAddress())
	if e != nil {
		return e
	}
	blastr := "ㄜ0:0"
	if febls != nil {
		blastr = febls.Hacash.ToFinString()
	}
	if febls == nil || febls.Hacash.LessThan(txfee) {
		// 余额不足以支付手续费
		return fmt.Errorf("diamond create tx fee address %s balance need not less than %s but got %s.", newtx.GetAddress(), txfee.ToFinString(), blastr)
	}
	//return nil
	return p.blockchain.ValidateDiamondCreateAction(act)
}
