package localcpu

import "github.com/hacash/core/sys"

type LocalCPUPowMasterConfig struct {
	Concurrent uint32 // 并发挖矿

}

func NewEmptyLocalCPUPowMasterConfig() *LocalCPUPowMasterConfig {
	cnf := &LocalCPUPowMasterConfig{
		Concurrent: 1,
	}
	return cnf
}

//////////////////////////////////////////////////

func NewLocalCPUPowMasterConfig(cnffile *sys.Inicnf) *LocalCPUPowMasterConfig {
	cnf := NewEmptyLocalCPUPowMasterConfig()

	return cnf

}
