package miner

import (
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/sys"
	"os"
)

type MinerConfig struct {
	Rewards fields.Address
}

func NewEmptyMinerConfig() *MinerConfig {
	cnf := &MinerConfig{}
	return cnf
}

//////////////////////////////////////////////////

func NewMinerConfig(cnffile *sys.Inicnf) *MinerConfig {
	cnf := NewEmptyMinerConfig()

	section := cnffile.Section("miner")
	rwdstr := section.Key("rewards").MustString("1AVRuFXNFi3rdMrPH4hdqSgFrEBnWisWaS")
	addr, err := fields.CheckReadableAddress(rwdstr)
	if err == nil {
		cnf.Rewards = *addr
	} else {
		fmt.Println(err)
		os.Exit(0)
	}
	return cnf
}
