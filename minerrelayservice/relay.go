package minerrelayservice

import (
	"github.com/hacash/chain/leveldb"
	"github.com/hacash/miner/message"
	"net"
	"net/http"
	"sync"
)

type RelayService struct {
	config *MinerRelayServiceConfig

	service_tcp *net.TCPConn

	changelock sync.Mutex

	allconns map[uint64]*ConnClient // 全部 TCP 连接

	oldprevBlockStuff  *message.MsgPendingMiningBlockStuff // 上一个挖掘的区块消息
	penddingBlockStuff *message.MsgPendingMiningBlockStuff // 当前正在挖掘的区块消息
	//successMintCh    chan interfaces.Block               // 当前正确挖掘区块的返回

	// routes
	queryRoutes     map[string]func(*http.Request, http.ResponseWriter, []byte)
	createRoutes    map[string]func(*http.Request, http.ResponseWriter, []byte)
	submitRoutes    map[string]func(*http.Request, http.ResponseWriter, []byte)
	operateRoutes   map[string]func(*http.Request, http.ResponseWriter, []byte)
	calculateRoutes map[string]func(*http.Request, http.ResponseWriter, []byte)

	// data
	ldb *leveldb.DB

	userMiningResultStoreAutoIdxMutex sync.Mutex
	userMiningResultStoreAutoIdx      uint64
}

func NewRelayService(cnf *MinerRelayServiceConfig) *RelayService {
	return &RelayService{
		config:                       cnf,
		service_tcp:                  nil,
		allconns:                     make(map[uint64]*ConnClient),
		oldprevBlockStuff:            nil,
		penddingBlockStuff:           nil,
		queryRoutes:                  make(map[string]func(*http.Request, http.ResponseWriter, []byte)),
		createRoutes:                 make(map[string]func(*http.Request, http.ResponseWriter, []byte)),
		submitRoutes:                 make(map[string]func(*http.Request, http.ResponseWriter, []byte)),
		operateRoutes:                make(map[string]func(*http.Request, http.ResponseWriter, []byte)),
		calculateRoutes:              make(map[string]func(*http.Request, http.ResponseWriter, []byte)),
		ldb:                          nil,
		userMiningResultStoreAutoIdx: 0,
	}
}

// 新的挖矿数据到来
func (r *RelayService) updateNewBlockStuff(newstf *message.MsgPendingMiningBlockStuff) {
	r.oldprevBlockStuff = r.penddingBlockStuff // 保存上一个
	r.penddingBlockStuff = newstf              // 更新最新的
	// 储存至磁盘
	go r.saveMiningBlockStuffToStore(newstf)
}

// 找出 stuff 通过 区块高度
func (r *RelayService) checkoutMiningStuff(blkhei uint64) *message.MsgPendingMiningBlockStuff {
	if r.oldprevBlockStuff != nil {
		if r.oldprevBlockStuff.BlockHeadMeta.GetHeight() == blkhei {
			return r.oldprevBlockStuff
		}
	}
	if r.penddingBlockStuff != nil {
		if r.penddingBlockStuff.BlockHeadMeta.GetHeight() == blkhei {
			return r.penddingBlockStuff
		}
	}
	// 不存在
	return nil
}

func (r *RelayService) Start() {

	r.initStore()

	go r.startListen() // 启动 server 服务端

	go r.startHttpApiService() // 启动 http api 服务

	go r.loop() // 启动 loop

	go r.connectToService() // 连接至服务器

}
