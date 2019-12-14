package memtxpool

import (
	"github.com/hacash/core/actions"
	"github.com/hacash/core/interfaces"
)

type TxItem struct {
	tx        interfaces.Transaction
	size      uint32
	feepurity uint64

	prev *TxItem
	next *TxItem

	diamond *actions.Action_4_DiamondCreate
}
