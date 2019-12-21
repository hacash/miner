package console

import "github.com/hacash/core/sys"

type MinerConsoleConfig struct {
	HttpListenPort int
}

func NewEmptyMinerConsoleConfig() *MinerConsoleConfig {
	cnf := &MinerConsoleConfig{
		HttpListenPort: 3340,
	}
	return cnf
}

func NewMinerConsoleConfig(cnffile *sys.Inicnf) *MinerConsoleConfig {
	cnf := NewEmptyMinerConsoleConfig()

	mpsec := cnffile.Section("minerpool")
	cnf.HttpListenPort = mpsec.Key("console_http_port").MustInt(3340)

	return cnf
}
