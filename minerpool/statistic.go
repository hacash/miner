package minerpool

import (
	"github.com/hacash/core/fields"
	"github.com/hacash/mint/difficulty"
	"math/big"
)

// 统计算力
func (a *Account) addPowWorth(hash fields.Hash) {
	a.change.Lock()
	defer a.change.Unlock()

	//fmt.Println("addPowWorth", a, hash.ToHex())

	val := CalculateHashWorth(hash)
	a.realtimePowWorth = new(big.Int).Add(a.realtimePowWorth, val)
}

///////////////////////////////////////////

// 计算哈希价值
func CalculateHashWorth(hash []byte) *big.Int {
	mulnum := big.NewInt(2)
	worth := big.NewInt(2)
	prezorenum := 0
	wbits := difficulty.BytesToBits(hash)
	for i, v := range wbits {
		if v != 0 {
			prezorenum = i
			break
		}
	}
	//
	for i := 0; i < prezorenum; i++ {
		worth = worth.Mul(worth, mulnum)
	}
	return worth
}
