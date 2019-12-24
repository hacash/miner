package memtxpool

import (
	"github.com/hacash/core/actions"
)

func (p *MemTxPool) checkDiamondCreate(act *actions.Action_4_DiamondCreate) error {
	return p.blockchain.ValidateDiamondCreateAction(act)
}
