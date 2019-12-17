package miner

import "github.com/hacash/core/sys"

type MinerConfig struct {
}

func NewEmptyMinerConfig() *MinerConfig {
	cnf := &MinerConfig{}
	return cnf
}

//////////////////////////////////////////////////

func NewMinerConfig(cnffile *sys.Inicnf) *MinerConfig {
	cnf := NewEmptyMinerConfig()

	return cnf

}
