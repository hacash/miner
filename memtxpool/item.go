package memtxpool

import (
	"github.com/hacash/core/actions"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfacev2"
)

type TxItem struct {
	tx        interfacev2.Transaction
	hash      fields.Hash
	size      uint32
	feepurity uint64

	next *TxItem
	prev *TxItem

	diamond *actions.Action_4_DiamondCreate
}

func (p *TxItem) GetTx() interfacev2.Transaction {
	return p.tx
}

func (p *TxItem) GetNext() *TxItem {
	return p.next
}
