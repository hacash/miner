package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/account"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfacev2"
	"github.com/hacash/core/transactions"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

/**

测试借贷相关功能，流程

0. 启动比特币转移日志服务：

	cd mint/run/btcmovelogs
	go run main.go

1. 删除 test/data1 目录，启动：

	go build -o test/test1    miner/run/main/main.go && ./test/test1    test1.ini

2. 按顺序提交下面的交易，创建基础数据：


*/

/*
	acc1 := account.CreateAccountByPassword("123456") // 1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9
	acc2 := account.CreateAccountByPassword("1234567") // 18q1X1gpUAi97rHeT7NAriKS6ZHP1TVYjj
	acc3 := account.CreateAccountByPassword("12345678") // 1EDUeK8NAjrgYhgDFv9NJecn8dNyJJsu3y
	acc4 := account.CreateAccountByPassword("123456789") // 1P6DHQYjP6WygqTCzwXpwo7TxWqhA1SgVY
*/

// Create basic data and submit
func Test_create_base_data_submit1(t *testing.T) {

	// Create bitcoin transfer transaction
	post_btcmove_tx()

}

// Test bitcoin system lending
func Test_bitcoin_system_lending_create(t *testing.T) {

	hash15, _ := hex.DecodeString("130dd68299cf6d2bd68299cf6d2b2b")

	amt1 := fields.NewAmountSmall(8, 248)
	amt2 := fields.NewAmountSmall(1, 248)

	act1 := &actions.Action_17_BitcoinsSystemLendingCreate{
		LendingID:                hash15,
		MortgageBitcoinPortion:   10,
		LoanTotalAmount:          *amt1,
		PreBurningInterestAmount: *amt2,
	}

	post_1MzNY_tx_for_action(act1, nil)

}

// Test bitcoin system lending and redemption
func Test_bitcoin_system_lending_ransom(t *testing.T) {

	hash15, _ := hex.DecodeString("130dd68299cf6d2bd68299cf6d2b2b")
	addr1, _ := fields.CheckReadableAddress("1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9") // 私有赎回
	// addr1, _:= fields.CheckReadableAddress("1EDUeK8NAjrgYhgDFv9NJecn8dNyJJsu3y") // 公共赎回

	amt1 := fields.NewAmountSmall(8, 248)

	act1 := &actions.Action_18_BitcoinsSystemLendingRansom{
		LendingID:    hash15,
		RansomAmount: *amt1,
	}

	post_tx_for_action(act1, addr1.ToReadable(), nil)

}

// Create basic data and submit
func Test_create_base_data_submit(t *testing.T) {

	// Create diamond
	post_diamond_tx("AAABBB", 1)
	time.Sleep(time.Second * 11)
	post_diamond_tx("XYXYXY", 2)
	time.Sleep(time.Second * 11)
	post_diamond_tx("SSSNNN", 3)
	time.Sleep(time.Second * 11)
	post_diamond_tx("WTUVSB", 4)

	// Create bitcoin transfer transaction
	post_btcmove_tx()
	// First transfer
	time.Sleep(time.Second * 30)
	post_hactrs_tx(33, "1EDUeK8NAjrgYhgDFv9NJecn8dNyJJsu3y")

}

////////////////////////////////////////////////////////

// Test HAC transfer
func Test_hacash_trs(t *testing.T) {
	post_hactrs_tx(10, "1EDUeK8NAjrgYhgDFv9NJecn8dNyJJsu3y")
	//post_hactrs_tx(33, "18q1X1gpUAi97rHeT7NAriKS6ZHP1TVYjj")
}

// Test diamond system lending, test redemption
func Test_syslend_diamond_lending_ransom_loop(t *testing.T) {

	for i := 0; i < 1; i++ {

		syslend_diamond_lending_ransom()

		fmt.Println(time.Now().Unix(), "--------")

		time.Sleep(time.Second * 3)
	}
}

// Test diamond system lending, test redemption
func syslend_diamond_lending_ransom() {

	hash14, _ := hex.DecodeString("130dd68299cf6d2bd68299cf6d2b")
	// addr1, _:= fields.CheckReadableAddress("1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9") // 私有赎回
	addr1, _ := fields.CheckReadableAddress("1EDUeK8NAjrgYhgDFv9NJecn8dNyJJsu3y") // 公共赎回
	act1 := &actions.Action_16_DiamondsSystemLendingRansom{
		LendingID: hash14,
		RansomAmount: fields.Amount{
			248,
			1,
			[]byte{17},
		},
	}
	post_tx_for_action(act1, addr1.ToReadable(), nil)

}

// Test diamond system lending
func Test_syslend_diamond_lending(t *testing.T) {

	hash14, _ := hex.DecodeString("130dd68299cf6d2bd68299cf6d2b")
	//addr1, _ := fields.CheckReadableAddress("1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9")
	act1 := &actions.Action_15_DiamondsSystemLendingCreate{
		LendingID: hash14,
		MortgageDiamondList: fields.DiamondListMaxLen200{
			2,
			[]fields.DiamondName{[]byte("AAABBB"), []byte("XYXYXY")},
		},
		LoanTotalAmount: fields.Amount{
			248,
			1,
			[]byte{16},
		},
		BorrowPeriod: 20,
	}

	post_tx_for_action(act1, "1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9", nil)

}

// Test bitcoin transfer
func Test_satoshi_trs1(t *testing.T) {
	toaddr, _ := fields.CheckReadableAddress("1EDUeK8NAjrgYhgDFv9NJecn8dNyJJsu3y")
	act1 := &actions.Action_8_SimpleSatoshiTransfer{
		ToAddress: *toaddr,
		Amount:    200,
	}
	post_1MzNY_tx_for_action(act1, nil)
}
func Test_satoshi_trs2(t *testing.T) {
	addr1, _ := fields.CheckReadableAddress("1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9")
	addr2, _ := fields.CheckReadableAddress("1EDUeK8NAjrgYhgDFv9NJecn8dNyJJsu3y")
	act1 := &actions.Action_8_SimpleSatoshiTransfer{
		ToAddress: *addr1,
		Amount:    201,
	}
	post_tx_for_action(act1, addr2.ToReadable(), nil)
}

// Test inter user lending, test lender seizure
func Test_users_lending_clear(t *testing.T) {

	hash17, _ := hex.DecodeString("530dd68299cf6d2bd68299cf6d2b2bd682")
	ransomAmt := fields.Amount{
		0,
		0,
		[]byte{},
	}

	// redeem
	act1 := &actions.Action_20_UsersLendingRansom{
		LendingID:    hash17,
		RansomAmount: ransomAmt,
	}

	post_tx_for_action(act1, "1EDUeK8NAjrgYhgDFv9NJecn8dNyJJsu3y", nil)
}

// Test inter user lending and self redemption
func Test_users_lending_ransom(t *testing.T) {

	hash17, _ := hex.DecodeString("530dd68299cf6d2bd68299cf6d2b2bd682")

	act1 := &actions.Action_20_UsersLendingRansom{
		LendingID: hash17,
		RansomAmount: fields.Amount{
			248,
			1,
			[]byte{17},
		},
	}

	post_1MzNY_tx_for_action(act1, nil)

}

// Test inter user lending and public redemption
func Test_users_lending_public_ransom(t *testing.T) {

	hash17, _ := hex.DecodeString("530dd68299cf6d2bd68299cf6d2b2bd682")

	act1 := &actions.Action_20_UsersLendingRansom{
		LendingID: hash17,
		RansomAmount: fields.Amount{
			248,
			1,
			[]byte{17},
		},
	}

	post_tx_for_action(act1, "18q1X1gpUAi97rHeT7NAriKS6ZHP1TVYjj", nil)

}

// Test inter user lending
func Test_users_lending(t *testing.T) {

	hash17, _ := hex.DecodeString("530dd68299cf6d2bd68299cf6d2b2bd682")

	addr1, _ := fields.CheckReadableAddress("1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9")
	addr2, _ := fields.CheckReadableAddress("1EDUeK8NAjrgYhgDFv9NJecn8dNyJJsu3y")

	act1 := &actions.Action_19_UsersLendingCreate{
		LendingID:               hash17,
		IsRedemptionOvertime:    0,
		IsPublicRedeemable:      1,
		AgreedExpireBlockHeight: 510,
		MortgagorAddress:        *addr1,
		LenderAddress:           *addr2,
		MortgageBitcoin: fields.SatoshiVariation{
			1,
			50000,
		},
		MortgageDiamondList: fields.DiamondListMaxLen200{
			2,
			[]fields.DiamondName{[]byte("XYXYXY"), []byte("WTUVSB")},
		},
		//MortgageDiamondList: fields.DiamondListMaxLen200{
		//	0,
		//	[]fields.Bytes6{},
		//},
		LoanTotalAmount: fields.Amount{
			248,
			1,
			[]byte{15},
		},
		AgreedRedemptionAmount: fields.Amount{
			248,
			1,
			[]byte{17},
		},
		PreBurningInterestAmount: fields.Amount{
			246,
			1,
			[]byte{15},
		},
	}

	post_1MzNY_tx_for_action(act1, nil)

}

// Test diamond transfer
func Test_diamove(t *testing.T) {

	addr2, _ := fields.CheckReadableAddress("1EDUeK8NAjrgYhgDFv9NJecn8dNyJJsu3y")

	act1 := &actions.Action_5_DiamondTransfer{
		Diamond:   []byte("XYXYXY"),
		ToAddress: *addr2,
	}

	post_1MzNY_tx_for_action(act1, nil)

}

////////////////////////////////////////////////////////////////

// Create bitcoin transaction
func post_btcmove_tx() {

	mainaddr, _ := fields.CheckReadableAddress("1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9")
	hash32, _ := hex.DecodeString("8deb5180a3388fee4991674c62705041616980e76288a8888b65530e41ccf90d")

	// Create bitcoin transfer
	act1 := &actions.Action_34_SatoshiGenesis{
		TransferNo:               1,
		BitcoinBlockHeight:       1001,
		BitcoinBlockTimestamp:    1596702752,
		BitcoinEffectiveGenesis:  0,
		BitcoinQuantity:          1,
		AdditionalTotalHacAmount: 1048576,
		OriginAddress:            *mainaddr,
		BitcoinTransferHash:      hash32,
	}

	post_1MzNY_tx_for_action(act1, nil)

}

// Create a diamond transaction
func post_diamond_tx(diamond string, number uint32) {

	mainaddr, _ := fields.CheckReadableAddress("1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9")
	hash8, _ := hex.DecodeString("530dd68299cf6d2b")
	hash32, _ := hex.DecodeString("000000000e8ca4376218601120e12b6724a8c174087b9614530dd68299cf6d2b")

	// Create diamond
	act1 := &actions.Action_4_DiamondCreate{
		Diamond:       fields.DiamondName(diamond),
		Number:        fields.DiamondNumber(number),
		PrevHash:      hash32,
		Nonce:         hash8,
		Address:       *mainaddr,
		CustomMessage: hash32,
	}
	post_1MzNY_tx_for_action(act1, nil)

}

// Basic transfer
func post_hactrs_tx(hacnum int64, address string) {

	addr2, _ := fields.CheckReadableAddress(address)
	amt1 := fields.NewAmountByUnitMei(hacnum)
	// Create diamond
	act1 := &actions.Action_1_SimpleToTransfer{
		ToAddress: *addr2,
		Amount:    *amt1,
	}
	post_1MzNY_tx_for_action(act1, nil)

}

// Basic transfer

// Create basic data and submit
func post_1MzNY_tx_for_action(act1 interfacev2.Action, accs []account.Account) {
	post_tx_for_action(act1, "1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9", nil)
}

// Create basic data and submit
func post_tx_for_action(act1 interfacev2.Action, mainAddress string, accs []account.Account) {

	// tx
	feeamt, _ := fields.NewAmountFromFinString("ㄜ1:248")
	mainaddr, _ := fields.CheckReadableAddress(mainAddress)
	tx, _ := transactions.NewEmptyTransaction_2_Simple(*mainaddr)
	tx.Fee = *feeamt
	tx.Timestamp = 1618839282

	tx.AppendAction(act1)

	// autograph
	acc1 := account.CreateAccountByPassword("123456")    // 1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9
	acc2 := account.CreateAccountByPassword("1234567")   // 18q1X1gpUAi97rHeT7NAriKS6ZHP1TVYjj
	acc3 := account.CreateAccountByPassword("12345678")  // 1EDUeK8NAjrgYhgDFv9NJecn8dNyJJsu3y
	acc4 := account.CreateAccountByPassword("123456789") // 1P6DHQYjP6WygqTCzwXpwo7TxWqhA1SgVY
	addrPrivateKeys := map[string][]byte{}
	addrPrivateKeys[string(acc1.Address)] = acc1.PrivateKey
	addrPrivateKeys[string(acc2.Address)] = acc2.PrivateKey
	addrPrivateKeys[string(acc3.Address)] = acc3.PrivateKey
	addrPrivateKeys[string(acc4.Address)] = acc4.PrivateKey
	for _, v := range accs {
		addrPrivateKeys[string(v.Address)] = v.PrivateKey
	}
	tx.FillNeedSigns(addrPrivateKeys, nil)

	// serialize
	txbody, _ := tx.Serialize()
	fmt.Println("tx body:", hex.EncodeToString(txbody))

	// Submit transaction
	postbts := bytes.NewBuffer([]byte{0, 0, 0, 1})
	postbts.Write(txbody)
	resp, e3 := doBytesPost("http://127.0.0.1:33381/operate", postbts.Bytes())
	fmt.Println(string(resp), e3)
}

// body提交二进制数据
func doBytesPost(url string, data []byte) ([]byte, error) {

	body := bytes.NewReader(data)
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		fmt.Printf("http.NewRequest,[err=%s][url=%s]", err, url)
		return []byte(""), err
	}
	request.Header.Set("Connection", "Keep-Alive")
	var resp *http.Response
	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		fmt.Printf("http.Do failed,[err=%s][url=%s]", err, url)
		return []byte(""), err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("http.Do failed,[err=%s][url=%s]", err, url)
	}
	return b, err
}
