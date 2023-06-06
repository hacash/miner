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
	Continued  bool // Continuous diamond mining
	// Automatic bidding
	AutoBid                bool              // Open or not
	AutoCheckInterval      float64           // Bidding check interval, minimum 0.1 seconds
	AutoBidMaxFee          *fields.Amount    // Highest quotation for a single diamond
	AutoBidMarginFee       *fields.Amount    // Increase range of single quotation
	AutoBidIgnoreAddresses []*fields.Address // Give up competing addresses
}

func NewEmptyDiamondMinerConfig() *DiamondMinerConfig {
	cnf := &DiamondMinerConfig{
		Supervene:              1,
		FeeAmount:              fields.NewAmountSmall(1, 246),
		FeeAccount:             nil,
		Rewards:                nil,
		Continued:              false,
		AutoBid:                false,
		AutoCheckInterval:      5,
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
		fmt.Println("[Diamond Miner Conf Error]", err)
		os.Exit(0)
	}
	password := cnfsection.Key("fee_password").MustString("")
	if password == "" {
		log.Fatal("[Diamond Miner Conf Error] fee password cannot be empty.")
		os.Exit(0)
	}
	feeamount, err2 := fields.NewAmountFromFinString(cnfsection.Key("fee_amount").MustString("ㄜ4:244"))
	if err2 == nil {
		cnf.FeeAmount = feeamount
	} else {
		fmt.Println("[Diamond Miner Conf Error] FeeAmount:", err)
		os.Exit(0)
	}
	cnf.Continued = cnfsection.Key("continued").MustBool(false) // Continuous mining
	// Automatic bidding
	cnf.AutoBid = cnfsection.Key("autobid").MustBool(false)
	if cnf.AutoBid {
		cnf.AutoCheckInterval = cnfsection.Key("autobid_check_interval").MustFloat64(5)
		if cnf.AutoCheckInterval < 0.1 {
			cnf.AutoCheckInterval = 0.1 // Check once every 0.1 seconds at most
		}
		autobidMaxFee, err3 := fields.NewAmountFromFinString(cnfsection.Key("autobid_fee_max").MustString("ㄜ10:250"))
		if err3 == nil {
			cnf.AutoBidMaxFee = autobidMaxFee
		} else {
			fmt.Println("[Diamond Miner Conf Error] AutobidMaxFee:", err)
			os.Exit(0)
		}
		autobidMarginFee, err4 := fields.NewAmountFromFinString(cnfsection.Key("autobid_fee_margin").MustString("ㄜ1:246"))
		if err4 == nil {
			cnf.AutoBidMarginFee = autobidMarginFee
		} else {
			fmt.Println("[Diamond Miner Conf Error] AutoBidMarginFee:", err)
			os.Exit(0)
		}
		if cnf.AutoBidMarginFee.LessThan(fields.NewAmountSmall(1, 244)) {
			fmt.Println("[Diamond Miner Conf Error] AutoBidMarginFee can not less than ㄜ1:244")
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
	// Private key
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
