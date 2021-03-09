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

type WorkClient struct {
	conn            *net.TCPConn
	workBlockHeight uint64
	pingtime        *time.Time
	setend          bool
	miningStartTime time.Time
}

func NewClient(conn *net.TCPConn) *WorkClient {
	cli := &WorkClient{
		conn:            conn,
		workBlockHeight: 0,
		pingtime:        nil,
		setend:          false,
	}
	cli.miningStartTime = time.Now()
	return cli
}

type MinerWorker struct {
	config *MinerWorkerConfig

	worker message.PowDeviceWorker

	miningOutputCh          chan message.PowMasterMsg
	immediateStartConnectCh chan bool

	clients        map[uint64]*WorkClient
	client         *WorkClient
	isInConnecting bool

	powerTotalCmx mapset.Set

	statusMutex sync.Mutex

	currentPowMasterMsg        *message.PowMasterMsg
	currentPowMasterCreateTime time.Time
}

func NewMinerWorker(cnf *MinerWorkerConfig) *MinerWorker {

	pool := &MinerWorker{
		config:                  cnf,
		client:                  nil,
		miningOutputCh:          make(chan message.PowMasterMsg, 2),
		immediateStartConnectCh: make(chan bool, 2),
		clients:                 map[uint64]*WorkClient{},
		powerTotalCmx:           mapset.NewSet(),
		isInConnecting:          false,
	}

	wkcnf := localcpu.NewEmptyLocalCPUPowMasterConfig()
	wkcnf.Concurrent = cnf.Concurrent
	wkcnf.ReturnPowerHash = true // 上报哈希最大值
	pool.worker = localcpu.NewLocalCPUPowMaster(wkcnf)

	return pool

}

func (p *MinerWorker) Start() {

	fmt.Printf("[Start] connect: %s, rewards: %s, supervene: %d. \n",
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

func (p *MinerWorker) pickTargetClient(blkhei uint64) *WorkClient {
	//fmt.Printf("pickTargetClient  <%d> ", blkhei)
	for h, v := range p.clients {
		//fmt.Printf("  %d  ", v.workBlockHeight)
		if h == blkhei {
			delete(p.clients, blkhei)
			return v
		}
	}
	return nil
}
