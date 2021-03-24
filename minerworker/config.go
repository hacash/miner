package minerworker

import (
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/sys"
	"net"
)

type MinerWorkerConfig struct {
	PoolAddress   *net.TCPAddr
	Rewards       fields.Address // 奖励地址
	Supervene     uint32         // CPU 并发挖矿
	IsReportPower bool           // 是否上报算力
}

func NewEmptyMinerPoolWorkerConfig() *MinerWorkerConfig {
	cnf := &MinerWorkerConfig{
		Supervene: 1,
	}
	return cnf
}

//////////////////////////////////////////////////

func NewMinerWorkerConfig(cnffile *sys.Inicnf) *MinerWorkerConfig {
	cnf := NewEmptyMinerPoolWorkerConfig()
	cnfsection := cnffile.Section("")
	// pool
	addr, err := net.ResolveTCPAddr("tcp", cnfsection.Key("pool").MustString(""))
	if err != nil {
		fmt.Println(err)
		panic("pool ip:port is error.")
	}
	cnf.PoolAddress = addr
	// rewards
	rwdaddr, e1 := fields.CheckReadableAddress(cnfsection.Key("rewards").MustString("1AVRuFXNFi3rdMrPH4hdqSgFrEBnWisWaS"))
	if e1 != nil {
		fmt.Println(e1)
		panic("reward address is error.")
	}
	cnf.Rewards = *rwdaddr
	// supervene
	cnf.Supervene = uint32(cnfsection.Key("supervene").MustUint(1))
	// IsReportPower
	cnf.IsReportPower = cnfsection.Key("not_report_power").MustBool(false) == false
	// ok
	return cnf
}
