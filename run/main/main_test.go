package main

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/blocks"
	"github.com/hacash/x16rs"
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

func Test_fxibug1(t *testing.T) {

	// bbb_181491
	str1 := "01000002c4f3005f8e6d7500000000010ad536b98a036b57607b1d61c963830d542b055e8854fa7abfdd8e9a9916fe14d91b95d5481fee57ff65105dcd153ef981b4484a317c82e3ca35d1000000011f03da2ddbdb26340000000052d06d40b44e0e38746f9e2188102ee48ff1c96bf801015368656e7a656e506f6f6c000000000000"

	data1, _ := hex.DecodeString(str1)
	bbb_181491, _, _ := blocks.ParseBlock(data1, 0)

	fmt.Println(bbb_181491.HashFresh().ToHex(), bbb_181491.GetTimestamp(), bbb_181491.GetPrevHash().ToHex(), bbb_181491.GetMrklRoot().ToHex())

}

func Test_fxibug2(t *testing.T) {

	// bbb_181496
	str1 := "01000002c4f8005f8e711a00000000043b11868f10a3e29e43e7b1409897891185a992a3cce27fe94530cd9a9916fe14d91b95d5481fee57ff65105dcd153ef981b4484a317c82e3ca35d1000000013274ed29dbdb26340000000052d06d40b44e0e38746f9e2188102ee48ff1c96bf801015368656e7a656e506f6f6c000000000000"

	data1, _ := hex.DecodeString(str1)
	bbb_181496, _, _ := blocks.ParseBlock(data1, 0)

	fmt.Println(bbb_181496.HashFresh().ToHex(), bbb_181496.GetTimestamp(), bbb_181496.GetPrevHash().ToHex(), bbb_181496.GetMrklRoot().ToHex())

}

func Test_fxibug3(t *testing.T) {

	// bbb_181455
	// str1 := "01000002c515005f8e9dfb00000000060c2ae20c7eb06676b2180115978ce72ddb5611fb6ef02a5a1da464642f7819cfb7431d4847e2e4136218b92f81aa6275746a0f66176e145a26e21f000000025440ec5bdbdb26340000000052d06d40b44e0e38746f9e2188102ee48ff1c96bf801015368656e7a656e506f6f6c00000000000002005f8e96a90052d06d40b44e0e38746f9e2188102ee48ff1c96bf4010500010004544d53584954004f880000000005a67b36d33127a673b0bbbaea2f127beee529da355b52447825d7384125ce2d5847f0950052d06d40b44e0e38746f9e2188102ee48ff1c96b737b17401f38a897be977ec34a673b6249299563c04187c898b03d856639cae900010307d18b73279b14848e84d48f97576c3e6432df190b29da94df8b916570fe96f8208f4937abd955d141f06255ba3a32f46dde0a0b397d5e59ca10a59ad59c92d644186d2ba268822db03edd32e691f5d15cbf943a535f11621c76b21ccf7754990000"

	str2 := "01000002c515005f8e9dfb00000000060c2ae20c7eb06676b2180115978ce72ddb5611fb6ef02a5a1da464642f7819cfb7431d4847e2e4136218b92f81aa6275746a0f66176e145a26e21f000000025440ec5bdbdb26340000000052d06d40b44e0e38746f9e2188102ee48ff1c96bf801015368656e7a656e506f6f6c00000000000002005f8e96a90052d06d40b44e0e38746f9e2188102ee48ff1c96bf4010500010004544d53584954004f880000000005a67b36d33127a673b0bbbaea2f127beee529da355b52447825d7384125ce2d5847f0950052d06d40b44e0e38746f9e2188102ee48ff1c96b652d5d07df0d3a0ea04f1c21594489fb69e1ff32472fff8d502677684721533000010307d18b73279b14848e84d48f97576c3e6432df190b29da94df8b916570fe96f8208f4937abd955d141f06255ba3a32f46dde0a0b397d5e59ca10a59ad59c92d644186d2ba268822db03edd32e691f5d15cbf943a535f11621c76b21ccf7754990000"

	data1, _ := hex.DecodeString(str2)

	bbb_181455, _, _ := blocks.ParseBlock(data1, 0)

	trs := bbb_181455.GetTransactions()

	//fmt.Println(trs[0].Hash().ToHex())
	fmt.Println(trs[1].Hash().ToHex())
	fmt.Println(trs[1].GetFee().ToFinString())
	fmt.Println(trs[1].GetAddress().ToReadable())

	act := trs[1].GetActions()
	fmt.Println(act[0].Kind())

	diacrt := act[0].(*actions.Action_4_DiamondCreate)
	fmt.Println(string(diacrt.Diamond))
	fmt.Println(diacrt.Number)
	fmt.Println(hex.EncodeToString(diacrt.PrevHash))
	fmt.Println(hex.EncodeToString(diacrt.Nonce))
	fmt.Println(diacrt.Address.ToReadable())
	fmt.Println(hex.EncodeToString(diacrt.CustomMessage))
	fmt.Println("-----------------")

	fmt.Println(x16rs.Diamond(uint32(diacrt.Number), diacrt.PrevHash, diacrt.Nonce, diacrt.Address, diacrt.CustomMessage))

	fmt.Println("-----------------")
	blkmrklhx := blocks.CalculateMrklRoot(trs)

	fmt.Println(blkmrklhx.ToHex())
	fmt.Println(bbb_181455.GetMrklRoot().ToHex())

	fmt.Println(bbb_181455.HashFresh().ToHex(), bbb_181455.GetTimestamp(), bbb_181455.GetPrevHash().ToHex())

}
