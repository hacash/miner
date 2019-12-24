package diamondminer

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/account"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/sys"
	"log"
	"os"
)

type DiamondMinerConfig struct {
	Supervene  int
	FeeAmount  *fields.Amount
	FeeAccount *account.Account
	Rewards    fields.Address
}

func NewEmptyDiamondMinerConfig() *DiamondMinerConfig {
	cnf := &DiamondMinerConfig{
		Supervene:  1,
		FeeAmount:  fields.NewAmountSmall(4, 244),
		FeeAccount: nil,
		Rewards:    nil,
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
