package miningpool

import "github.com/hacash/core/sys"

type MinerPoolConfig struct {
}

func NewEmptyMinerPoolConfig() *MinerPoolConfig {
	cnf := &MinerPoolConfig{}
	return cnf
}

//////////////////////////////////////////////////

func NewMinerPoolConfig(cnffile *sys.Inicnf) *MinerPoolConfig {
	cnf := NewEmptyMinerPoolConfig()

	return cnf

}
