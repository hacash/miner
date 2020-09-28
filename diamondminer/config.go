package diamondminer

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/account"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/sys"
	"log"
	"os"
	"strings"
)

type DiamondMinerConfig struct {
	Supervene  int
	FeeAmount  *fields.Amount
	FeeAccount *account.Account
	Rewards    fields.Address
	// 自动竞价
	AutoBid                bool              // 是否开启
	AutoBidMaxFee          *fields.Amount    // 单枚钻石最高报价
	AutoBidMarginFee       *fields.Amount    // 单次报价提高幅度
	AutoBidIgnoreAddresses []*fields.Address // 放弃与之竞争的地址
}

func NewEmptyDiamondMinerConfig() *DiamondMinerConfig {
	cnf := &DiamondMinerConfig{
		Supervene:              1,
		FeeAmount:              fields.NewAmountSmall(1, 246),
		FeeAccount:             nil,
		Rewards:                nil,
		AutoBid:                false,
		AutoBidMaxFee:          fields.NewAmountSmall(10, 250), // 1000枚
		AutoBidMarginFee:       fields.NewAmountSmall(1, 246),
		AutoBidIgnoreAddresses: make([]*fields.Address, 0),
	}
	return cnf
}

//////////////////////////////////////////////////

func NewDiamondMinerConfig(cnffile *sys.Inicnf) *DiamondMinerConfig {
	cnf := NewEmptyDiamondMinerConfig()
	cnfsection := cnffile.Section("diamondminer")

	cnf.Supervene = cnfsection.Key("supervene").MustInt(1)
	rwdstr := cnfsection.Key("rewards").MustString("1AVRuFXNFi3rdMrPH4hdqSgFrEBnWisWaS")
	addr, err := fields.CheckReadableAddress(rwdstr)
	if err == nil {
		cnf.Rewards = *addr
	} else {
		fmt.Println("[Diamond Miner Config Error]", err)
		os.Exit(0)
	}
	password := cnfsection.Key("fee_password").MustString("")
	if password == "" {
		log.Fatal("[Diamond Miner Config Error] fee password cannot be empty.")
		os.Exit(0)
	}
	feeamount, err2 := fields.NewAmountFromFinString(cnfsection.Key("fee_amount").MustString("ㄜ4:244"))
	if err2 == nil {
		cnf.FeeAmount = feeamount
	} else {
		fmt.Println("[Diamond Miner Config Error] FeeAmount:", err)
		os.Exit(0)
	}
	// 自动竞价
	autobid := cnfsection.Key("autobid").MustString("false")
	if strings.Compare(autobid, "true") == 0 {
		cnf.AutoBid = true
	} else {
		cnf.AutoBid = false
	}
	if cnf.AutoBid {
		autobidMaxFee, err3 := fields.NewAmountFromFinString(cnfsection.Key("autobid_fee_max").MustString("ㄜ10:250"))
		if err3 == nil {
			cnf.AutoBidMaxFee = autobidMaxFee
		} else {
			fmt.Println("[Diamond Miner Config Error] AutobidMaxFee:", err)
			os.Exit(0)
		}
		autobidMarginFee, err4 := fields.NewAmountFromFinString(cnfsection.Key("autobid_fee_margin").MustString("ㄜ1:246"))
		if err4 == nil {
			cnf.AutoBidMarginFee = autobidMarginFee
		} else {
			fmt.Println("[Diamond Miner Config Error] AutoBidMarginFee:", err)
			os.Exit(0)
		}
		if cnf.AutoBidMarginFee.LessThan(fields.NewAmountSmall(1, 244)) {
			fmt.Println("[Diamond Miner Config Error] AutoBidMarginFee can not less than ㄜ1:244")
			os.Exit(0)
		}
		iasstrs := strings.Split(cnfsection.Key("autobid_ignore_addresses").MustString(""), ",")
		for _, iasstr := range iasstrs { // 忽略、放弃竞争的地址
			iasstr = strings.Replace(iasstr, " ", "", -1)
			if iasstr == "" {
				continue
			}
			addr, err := fields.CheckReadableAddress(iasstr)
			if err == nil {
				cnf.AutoBidIgnoreAddresses = append(cnf.AutoBidIgnoreAddresses, addr)
			} else {
				fmt.Println("AutoBidIgnoreAddresses:", err)
				os.Exit(0)
			}
		}
	}
	// 私钥
	var privkey []byte = nil
	if len(password) == 64 {
		key, err := hex.DecodeString(password)
		if err == nil {
			privkey = key
		}
	}
	var fee_acc *account.Account = nil
	if privkey != nil {
		fee_acc, err = account.GetAccountByPriviteKey(privkey)
	} else {
		fee_acc = account.CreateAccountByPassword(password)
	}
	if err != nil {
		panic(err)
	}
	cnf.FeeAccount = fee_acc
	// OK
	return cnf
}
