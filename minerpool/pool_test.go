package minerpool

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/chain/leveldb"
	"testing"
)

func Test_t1(t *testing.T) {

	hashstrs := []string{
		"000000000000000000000000000000000000000000421d81934fc36724a894c5",
		"0000000000000000000fdcf6268b0ba843221f0a50421d81934fc36724a894c5",
		"000000000292cf0d9c7dd077629b34f6e11d4b586ac837f281bae858269c52c8",
		"0002004412345678f6e11d4b586ac837f281bae85221f0a50421d81d077629b3",
	}
	for _, v := range hashstrs {
		hash, _ := hex.DecodeString(v)
		worth := CalculateHashWorth(hash)
		fmt.Println(worth.String(), worth.Bytes())
	}

	db, err := leveldb.OpenFile("/home/shiqiujie/Desktop/Hacash/go/src/github.com/hacash/miner/miningpool/testdata", nil)
	fmt.Println(err)

	err = db.Put([]byte("key"), []byte("value222"), nil)
	fmt.Println(err)

	fmt.Println(db.Get([]byte("key"), nil))

}
