package main

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/blocks"
	"testing"
)

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
