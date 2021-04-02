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
	cnf.TcpListenPort = cnfsection.Key("listen_port").MustInt(3351)
	cnf.MaxWorkerConnect = cnfsection.Key("max_connect").MustInt(200)
	// ok
	return cnf
}
