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
		config:      cnf,
		service_tcp: nil,
		allconns:    make(map[uint64]*ConnClient),
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

/*
// New mining data coming
func (r *RelayService) updateNewBlockStuff(newstf *interfaces2.PoWStuffOverallData) {
	r.oldprevBlockStuff = r.penddingBlockStuff // Save previous
	r.penddingBlockStuff = newstf              // Update the latest
	// Save to disk
	go r.saveMiningBlockStuffToStore(newstf)
}

// Find the height of the stuff passing through the block
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
	// non-existent
	return nil
}
*/

func (r *RelayService) Start() {

	r.initStore()

	r.startListen() // Start the server

	r.startHttpApiService() // Start HTTP API service

	r.connectToService() // Connect to server

	go r.loop() // Start loop

}
