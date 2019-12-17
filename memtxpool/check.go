package memtxpool

import (
	"fmt"
	"github.com/hacash/core/actions"
)

func (p *MemTxPool) checkDiamondCreate(newdiamond *actions.Action_4_DiamondCreate) error {
	last, err := p.blockchain.State().ReadLastestDiamond()
	if err != nil {
		return err
	}
	if uint32(newdiamond.Number) != uint32(last.Number)+1 {
		return fmt.Errorf("Diamond number error.")
	}
	if last.ContainBlockHash.Equal(newdiamond.PrevHash) != true {
		return fmt.Errorf("Diamond prev block hash error.")
	}
	return nil
}
