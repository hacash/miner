package minerpoolworker

import (
	"fmt"
	"github.com/hacash/chain/mapset"
	"github.com/hacash/miner/localcpu"
	"github.com/hacash/miner/localgpu"
	"github.com/hacash/miner/message"
	"net"
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

type MinerPoolWorker struct {
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

func NewMinerWorker(cnf *MinerWorkerConfig) *MinerPoolWorker {

	pool := &MinerPoolWorker{
		config:                  cnf,
		client:                  nil,
		miningOutputCh:          make(chan message.PowMasterMsg, 2),
		immediateStartConnectCh: make(chan bool, 2),
		clients:                 map[uint64]*WorkClient{},
		powerTotalCmx:           mapset.NewSet(),
		isInConnecting:          false,
	}
	// here change to GPU
	if cnf.GPU_Enable == true {
		wkcnf := localgpu.NewEmptyLocalGPUPowMasterConfig()
		wkcnf.Concurrent = cnf.Concurrent
		wkcnf.ReturnPowerHash = true // Maximum reported hash
		wkcnf.OpenclPath = cnf.GPU_OpenclPath
		wkcnf.PlatName = cnf.GPU_PlatformNameMatch
		wkcnf.GroupNum = cnf.GPU_GroupConcurrentNum
		wkcnf.GroupSize = cnf.GPU_GroupSize
		wkcnf.ItemLoop = cnf.GPU_ItemLoopNum
		wkcnf.EmptyFuncTest = cnf.GPU_EmptyFuncTest
		wkcnf.UseOneDeviceBuild = cnf.GPU_UseOneDeviceBuild
		pool.worker = localgpu.NewLocalGPUPowMaster(wkcnf)
	} else {
		wkcnf := localcpu.NewEmptyLocalCPUPowMasterConfig()
		wkcnf.Concurrent = cnf.Concurrent
		wkcnf.ReturnPowerHash = true // Maximum reported hash
		pool.worker = localcpu.NewLocalCPUPowMaster(wkcnf)
	}

	return pool

}

func (p *MinerPoolWorker) Start() {

	fmt.Printf("[Start] connect: %s, rewards: %s, supervene: %d. \n",
		p.config.PoolAddress.String(),
		p.config.Rewards.ToReadable(),
		p.config.Concurrent,
	)

	err := p.startConnect()
	if err != nil {
		fmt.Println("--------[ERROR]--------")
		fmt.Println(err)
		fmt.Println("--------[ERROR]--------")
		return
	}

	go p.loop()
}

func (p *MinerPoolWorker) pickTargetClient(blkhei uint64) *WorkClient {
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
