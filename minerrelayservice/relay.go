package minerrelayservice

import (
	"github.com/hacash/chain/leveldb"
	interfaces2 "github.com/hacash/miner/interfaces"
	"net"
	"net/http"
	"sync"
)

type RelayService struct {
	config *MinerRelayServiceConfig

	service_tcp *net.TCPConn

	changelock sync.Mutex

	allconns map[uint64]*ConnClient // All TCP connections

	//oldprevBlockStuff  *interfaces2.PoWStuffOverallData // Last mined block message
	penddingBlockStuff *interfaces2.PoWStuffOverallData // Currently mining block messages
	prevBlockStuffMaps sync.Map
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

	hashratepool interfaces2.HashratePool
}

func NewRelayService(cnf *MinerRelayServiceConfig) *RelayService {
	return &RelayService{
		config:             cnf,
		service_tcp:        nil,
		allconns:           make(map[uint64]*ConnClient),
		prevBlockStuffMaps: sync.Map{},
		//oldprevBlockStuff:            nil,
		//penddingBlockStuff:           nil,
		queryRoutes:                  make(map[string]func(*http.Request, http.ResponseWriter, []byte)),
		createRoutes:                 make(map[string]func(*http.Request, http.ResponseWriter, []byte)),
		submitRoutes:                 make(map[string]func(*http.Request, http.ResponseWriter, []byte)),
		operateRoutes:                make(map[string]func(*http.Request, http.ResponseWriter, []byte)),
		calculateRoutes:              make(map[string]func(*http.Request, http.ResponseWriter, []byte)),
		ldb:                          nil,
		userMiningResultStoreAutoIdx: 0,
	}
}

// New mining data coming
func (r *RelayService) updateNewBlockStuff(newstf *interfaces2.PoWStuffOverallData) {
	r.penddingBlockStuff = newstf // Update the latest
	r.prevBlockStuffMaps.Store(newstf.BlockHeadMeta.GetHeight(), newstf)
}

// Find the height of the stuff passing through the block
func (r *RelayService) checkoutMiningStuff(blkhei uint64) *interfaces2.PoWStuffOverallData {
	if r.penddingBlockStuff != nil {
		if r.penddingBlockStuff.BlockHeadMeta.GetHeight() == blkhei {
			return r.penddingBlockStuff
		}
	}
	var resobj *interfaces2.PoWStuffOverallData = nil
	sto, ok := r.prevBlockStuffMaps.Load(blkhei)
	if ok {
		resobj = sto.(*interfaces2.PoWStuffOverallData)
	}
	return resobj
}

func (r *RelayService) SetHashratePool(pool interfaces2.HashratePool) {
	r.hashratepool = pool
}

func (r *RelayService) Start() {

	r.initStore()

	r.startListen() // Start the server

	r.startHttpApiService() // Start HTTP API service

	r.connectToService() // Connect to server

	go r.loop() // Start loop

}
