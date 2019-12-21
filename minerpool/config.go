package minerpool

import (
	"github.com/hacash/core/sys"
	"path"
)

type MinerPoolConfig struct {
	Datadir           string
	TcpListenPort     int
	TcpConnectMaxSize uint
	FeePercentage     float64
}

func NewEmptyMinerPoolConfig() *MinerPoolConfig {
	cnf := &MinerPoolConfig{
		TcpListenPort:     3339,
		TcpConnectMaxSize: 200,
		FeePercentage:     0.2,
	}
	return cnf
}

//////////////////////////////////////////////////

func NewMinerPoolConfig(cnffile *sys.Inicnf) *MinerPoolConfig {
	cnf := NewEmptyMinerPoolConfig()
	cnfsection := cnffile.Section("minerpool")
	defdir := path.Join(path.Dir(cnffile.MustDataDir()), ".hacash_minerpool")
	cnf.Datadir = sys.AbsDir(cnfsection.Key("data_dir").MustString(defdir))
	cnf.TcpListenPort = cnfsection.Key("listen_port").MustInt(3339)
	cnf.TcpConnectMaxSize = cnfsection.Key("max_connect").MustUint(200)
	cnf.FeePercentage = cnfsection.Key("fee_percentage").MustFloat64(0.2)
	if cnf.FeePercentage >= 1 || cnf.FeePercentage < 0 {
		panic("fee_percentage value error.")
	}
	return cnf
}
