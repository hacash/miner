package miningpool

import (
	"github.com/hacash/core/sys"
	"path"
)

type MinerPoolConfig struct {
	Datadir           string
	TcpListenPort     int
	TcpConnectMaxSize uint
}

func NewEmptyMinerPoolConfig() *MinerPoolConfig {
	cnf := &MinerPoolConfig{
		TcpListenPort:     3339,
		TcpConnectMaxSize: 200,
	}
	return cnf
}

//////////////////////////////////////////////////

func NewMinerPoolConfig(cnffile *sys.Inicnf) *MinerPoolConfig {
	cnf := NewEmptyMinerPoolConfig()
	cnfsection := cnffile.Section("minerpool")
	defdir := path.Join(cnffile.MustDataDir(), "minerpool")
	cnf.Datadir = cnfsection.Key("data_dir").MustString(defdir)
	cnf.TcpListenPort = cnfsection.Key("listen_port").MustInt(3339)
	cnf.TcpConnectMaxSize = cnfsection.Key("max_connect").MustUint(200)
	return cnf
}
