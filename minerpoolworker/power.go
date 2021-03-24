package minerpoolworker

import (
	"github.com/hacash/mint/difficulty"
	"math/big"
)

func (p *MinerWorker) addPowerLogReturnShow(hxworth *big.Int) string {

	p.powerTotalCmx.Add(hxworth)
	if p.powerTotalCmx.Cardinality() > 24 {
		p.powerTotalCmx.Pop()
	}

	resworth := big.NewInt(0)
	for _, v := range p.powerTotalCmx.ToSlice() {
		num := v.(*big.Int)
		resworth = new(big.Int).Add(resworth, num)
	}

	resworth = new(big.Int).Div(resworth, big.NewInt(int64(p.powerTotalCmx.Cardinality())))

	return difficulty.ConvertPowPowerToShowFormat(resworth)
}
