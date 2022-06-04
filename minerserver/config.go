package minerserver

import (
	"github.com/hacash/core/sys"
)

type MinerServerConfig struct {
	TcpListenPort    int // TCP server listening port
	MaxWorkerConnect int // TCP Max connections
}

func NewEmptyMinerConfig() *MinerServerConfig {
	cnf := &MinerServerConfig{}
	return cnf
}

//////////////////////////////////////////////////

func NewMinerConfig(cnffile *sys.Inicnf) *MinerServerConfig {
	cnf := NewEmptyMinerConfig()

	section := cnffile.Section("minerserver")

	cnf.TcpListenPort = section.Key("listen_port").MustInt(3351)
	cnf.MaxWorkerConnect = section.Key("max_connect").MustInt(200)

	return cnf
}
