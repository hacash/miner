package memtxpool

import (
	"fmt"
	"github.com/hacash/core/account"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/transactions"
	"testing"
)

func Test_t1(t *testing.T) {

	grp := NewTxGroup()

	// txs
	txitem1 := createItemByFee("144:244")
	txitem2 := createItemByFee("148:248")
	txitem3 := createItemByFee("1486:247")
	fmt.Println(txitem1.tx.Size()/8, txitem1.tx.FeePurity())
	fmt.Println(txitem2.tx.Size()/8, txitem2.tx.FeePurity())
	fmt.Println(txitem3.tx.Size()/8, txitem3.tx.FeePurity())

	// do add
	grp.Add(txitem1)
	grp.Add(txitem2)
	grp.Add(txitem3)

	// for
	curitem := grp.Head
	for true {
		if curitem != nil {
			fmt.Println(curitem.feepurity)
			curitem = curitem.next
		} else {
			break
		}
	}

}

func createItemByFee(feestr string) *TxItem {

	// tx1
	addr1 := account.CreateNewRandomAccount().Address
	tx1, _ := transactions.NewEmptyTransaction_2_Simple(addr1)
	fee1, _ := fields.NewAmountFromFinString(feestr)
	tx1.Fee = *fee1

	txitem1 := &TxItem{
		tx:        tx1,
		hash:      tx1.Hash(),
		size:      tx1.Size(),
		feepurity: tx1.FeePurity(),
		diamond:   nil,
	}

	return txitem1
}
