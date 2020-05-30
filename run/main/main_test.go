package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/stores"
	"os"
	"testing"
)

func Test_showAllAmtByOneStateFile(t *testing.T) {

	// 遍历余额
	fr, e := os.Open("/home/shiqiujie/Desktop/Hacash/go/src/github.com/hacash/test/blk20200530.dat")
	if e != nil {
		fmt.Println(e.Error())
		os.Exit(0)
	}
	amtzore := fields.Amount{}
	itlen := (21 + 32)
	blkitem := bytes.Repeat([]byte{0}, itlen)
	blkii := int64(0)
	for {
		_, re := fr.ReadAt(blkitem, blkii*int64(itlen))
		if re != nil {
			fmt.Println(re.Error())
			os.Exit(0)
		}
		//fmt.Println(blkitem)
		addr := fields.Address(blkitem[:21])
		amt := stores.Balance{}
		_, be := amt.Parse(blkitem[21:], 0)
		if be != nil {
			fmt.Println(be.Error())
			os.Exit(0)
		}
		if !amt.Amount.Equal(&amtzore) {
			fmt.Println(addr.ToReadable(), amt.Amount.ToFinString())
		}

		blkii++
	}

}

func Test_t1(t *testing.T) {
	str1 := "010000000001005dfe0346000000077790ba2fcdeaef4a4299d9b667135bac577ce204dee8388f1b97f7e63ddba8b8dce81b2578e5de8c76efaf989c62b5f91505fd39adebcd3ee362fad10000000100000000fffffffe00000000e63c33a796b3032ce6b856f68fccf06608d9ed18f801012020202020202020202020000000000100"
	str2 := "010000000001005dfe0346000000077790ba2fcdeaef4a4299d9b667135bac577ce204dee8388f1b97f7e63ddba8b8dce81b2578e5de8c76efaf989c62b5f91505fd39adebcd3ee362fad10000000100000000fffffffe00000000e63c33a796b3032ce6b856f68fccf06608d9ed18f801012020202020202020202020000000000100"

	fmt.Println(str1 == str2)

	data1, _ := hex.DecodeString(str1)

	bbb, _, _ := blocks.ParseBlock(data1, 0)

	fmt.Println(bbb.GetMrklRoot().ToHex())

	trs := bbb.GetTransactions()
	fmt.Println(blocks.CalculateMrklRoot(trs).ToHex())
	fmt.Println(trs[0].Hash().ToHex())

}
