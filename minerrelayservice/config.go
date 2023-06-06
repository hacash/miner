package minerrelayservice

import (
	"fmt"
	"github.com/hacash/core/sys"
	"net"
	"os"
)

type MinerRelayServiceConfig struct {
	ServerAddress    *net.TCPAddr
	IsReportHashrate bool // Whether to report calculation force statistics
	IsAcceptHashrate bool // Whether to accept calculation force statistics

	ServerTcpListenPort int // TCP server listening port
	MaxWorkerConnect    int // TCP Max connections

	HttpApiListenPort int // HTTP API data interface service

	// Data storage
	StoreEnable          bool   // Storage on
	DataDir              string // catalogue
	SaveMiningBlockStuff bool   // Save mining block information
	SaveMiningHash       bool   // Whether to save the submitted block hash value
	SaveMiningNonce      bool   // Save submitted nonce value

}

func NewEmptyMinerRelayServiceConfig() *MinerRelayServiceConfig {
	cnf := &MinerRelayServiceConfig{}
	return cnf
}

//////////////////////////////////////////////////

func NewMinerRelayServiceConfig(cnffile *sys.Inicnf) *MinerRelayServiceConfig {
	cnf := NewEmptyMinerRelayServiceConfig()
	cnfsection := cnffile.Section("")
	// pool
	addr, err := net.ResolveTCPAddr("tcp", cnfsection.Key("server_connect").MustString(""))
	if err != nil {
		fmt.Println(err)
		panic("pool ip:port is error.")
	}
	cnf.ServerAddress = addr
	// IsReportHashrate  or  IsAcceptHashrate
	cnf.IsReportHashrate = cnfsection.Key("not_report_hashrate").MustBool(false) == false
	cnf.IsAcceptHashrate = cnfsection.Key("not_accept_hashrate").MustBool(false) == false
	// max
	cnf.MaxWorkerConnect = cnfsection.Key("max_connect").MustInt(200)
	cnf.ServerTcpListenPort = cnfsection.Key("server_listen_port").MustInt(19991)
	if cnf.ServerTcpListenPort == 0 {
		panic("Relay service:server listen port is zero")
	}

	cnf.HttpApiListenPort = cnfsection.Key("http_api_listen_port").MustInt(8080)
	// store
	storesection := cnffile.Section("store")
	cnf.StoreEnable = storesection.Key("enable").MustBool(false)
	if cnf.StoreEnable {
		cnf.DataDir = storesection.Key("data_dir").MustString("./hacash_relay_service_data")
		cnf.SaveMiningBlockStuff = storesection.Key("save_mining_block_stuff").MustBool(false)
		cnf.SaveMiningHash = storesection.Key("save_mining_hash").MustBool(false)
		cnf.SaveMiningNonce = storesection.Key("save_mining_nonce").MustBool(false)
		if !cnf.SaveMiningHash && !cnf.SaveMiningNonce {
			fmt.Println("[Miner Relay Service Conf] Error: SaveMiningHash and SaveMiningNonce cannot be both false!")
			os.Exit(0)
		}
	}
	// ok
	return cnf
}
