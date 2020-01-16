package minerworker

import (
	"fmt"
	"github.com/hacash/miner/localcpu"
	"github.com/hacash/miner/message"
	"net"
	"os"
	"time"
)


type Client struct {
	conn *net.TCPConn
	workBlockHeight uint64
	pingtime *time.Time
}

func NewClient(conn *net.TCPConn) *Client {
	return &Client{
		conn: conn,
		workBlockHeight: 0,
		pingtime: nil,
	}
}

type MinerWorker struct {
	config *MinerWorkerConfig

	worker *localcpu.LocalCPUPowMaster

	miningOutputCh          chan message.PowMasterMsg
	immediateStartConnectCh chan bool

	currentMiningStatusSuccess bool

	client *Client

}

func NewMinerWorker(cnf *MinerWorkerConfig) *MinerWorker {

	pool := &MinerWorker{
		currentMiningStatusSuccess: false,
		config:                     cnf,
		client:                       nil,
		miningOutputCh:             make(chan message.PowMasterMsg, 2),
		immediateStartConnectCh:    make(chan bool, 2),
	}

	wkcnf := localcpu.NewEmptyLocalCPUPowMasterConfig()
	wkcnf.Concurrent = cnf.Concurrent
	wkcnf.ReturnPowerHash = true // 上报哈希最大值
	pool.worker = localcpu.NewLocalCPUPowMaster(wkcnf)

	return pool

}

func (p *MinerWorker) Start() {

	err := p.startConnect()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	go p.loop()
}
