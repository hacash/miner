package minerworker

import (
	"fmt"
	"github.com/hacash/chain/mapset"
	"github.com/hacash/miner/localcpu"
	"github.com/hacash/miner/message"
	"net"
	"os"
	"sync"
	"time"
)


type Client struct {
	conn *net.TCPConn
	workBlockHeight uint64
	pingtime *time.Time
	setend bool
}

func NewClient(conn *net.TCPConn) *Client {
	return &Client{
		conn: conn,
		workBlockHeight: 0,
		pingtime: nil,
		setend: false,
	}
}

type MinerWorker struct {
	config *MinerWorkerConfig

	worker *localcpu.LocalCPUPowMaster

	miningOutputCh          chan message.PowMasterMsg
	immediateStartConnectCh chan bool

	clients mapset.Set
	client *Client

	statusMutex sync.Mutex
}

func NewMinerWorker(cnf *MinerWorkerConfig) *MinerWorker {

	pool := &MinerWorker{
		config:                     cnf,
		client:                       nil,
		miningOutputCh:             make(chan message.PowMasterMsg, 2),
		immediateStartConnectCh:    make(chan bool, 2),
		clients: mapset.NewSet(),
	}

	wkcnf := localcpu.NewEmptyLocalCPUPowMasterConfig()
	wkcnf.Concurrent = cnf.Concurrent
	wkcnf.ReturnPowerHash = true // 上报哈希最大值
	pool.worker = localcpu.NewLocalCPUPowMaster(wkcnf)

	return pool

}

func (p *MinerWorker) Start() {

	fmt.Printf("[Start] Connect: %s, rewards: %s, supervene: %d. \n",
		p.config.PoolAddress.String(),
		p.config.Rewards.ToReadable(),
		p.config.Concurrent,
	)

	err := p.startConnect()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	go p.loop()
}


func (p *MinerWorker) pickTargetClient( blkhei uint64 ) *Client {
	lists := p.clients.ToSlice()
	for _, v := range lists {
		if v.(*Client).workBlockHeight == blkhei {
			p.clients.Remove(v)
			return v.(*Client)
		}
	}
	return nil
}