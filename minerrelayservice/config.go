package minerrelayservice

import (
	"fmt"
	"github.com/hacash/core/sys"
	"net"
)

type MinerRelayServiceConfig struct {
	ServerAddress    *net.TCPAddr
	IsReportHashrate bool // 是否上报算力统计
	IsAcceptHashrate bool // 是否接受算力统计

	TcpListenPort    int // TCP server 监听端口
	MaxWorkerConnect int // TCP 最大连接数

	HttpApiListenPort int // http api 数据接口服务
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
	cnf.TcpListenPort = cnfsection.Key("server_listen_port").MustInt(0)
	cnf.HttpApiListenPort = cnfsection.Key("http_api_listen_port").MustInt(0)
	// ok
	return cnf
}
