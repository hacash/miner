package minerrelayservice

import (
	"fmt"
	"github.com/hacash/core/sys"
	"net"
	"os"
)

type MinerRelayServiceConfig struct {
	ServerAddress    *net.TCPAddr
	IsReportHashrate bool // 是否上报算力统计
	IsAcceptHashrate bool // 是否接受算力统计

	ServerTcpListenPort int // TCP server 监听端口
	MaxWorkerConnect    int // TCP 最大连接数

	HttpApiListenPort int // http api 数据接口服务

	// 数据储存
	StoreEnable          bool   // 储存开启
	DataDir              string // 目录
	SaveMiningBlockStuff bool   // 是否保存挖掘区块的信息
	SaveMiningHash       bool   // 是否保存提交的 block hash 值
	SaveMiningNonce      bool   // 是否保存提交的 nonce 值

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
	cnf.IsReportHashrate = cnfsection.Key("report_hashrate").MustBool(false)
	cnf.IsAcceptHashrate = cnfsection.Key("accept_hashrate").MustBool(true)
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
			fmt.Println("[Miner Relay Service Config] Error: SaveMiningHash and SaveMiningNonce cannot be both false!")
			os.Exit(0)
		}
	}
	// ok
	return cnf
}
