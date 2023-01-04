package main

import (
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfacev2"
	"github.com/hacash/core/sys"
	"github.com/hacash/miner/console"
	"github.com/hacash/miner/diamondminer"
	"github.com/hacash/miner/localcpu"
	"github.com/hacash/miner/memtxpool"
	"github.com/hacash/miner/miner"
	"github.com/hacash/miner/minerpool"
	"github.com/hacash/miner/minerserver"
	"github.com/hacash/mint"
	"github.com/hacash/mint/blockchain"
	"github.com/hacash/mint/blockchainv3"
	"github.com/hacash/node/backend"
	"github.com/hacash/node/p2pv2"
	deprecated "github.com/hacash/service/deprecated"
	rpc "github.com/hacash/service/rpc"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"
)

/**
# window and macOS need to build 'libx16rs_hash.a' by gcc in the first time!

export GOPATH=/media/yangjie/500GB/hacash/go

go build -o test/mainnet  miner/run/main/main.go && ./test/mainnet  mainnet.ini
go build -o test/test1    miner/run/main/main.go && ./test/test1    test1.ini
go build -o test/test2    miner/run/main/main.go && ./test/test2    test2.ini
go build -o test/test3    miner/run/main/main.go && ./test/test3    test3.ini
go build -o test/pcwallet pcwallet/main/main.go  && ./test/pcwallet
go build -o test/poolworkertest1 miner/run/minerpoolworker/main.go  && ./test/poolworkertest1 poolworkertest1.ini

*/

/**

编译发布版本：

go build -ldflags '-w -s' -o                   hacash_node_2022_11_27_01  miner/run/main/main.go
go build -ldflags '-w -s' -o      hacash_miner_pool_worker_2022_09_09_01  miner/run/minerpoolworker/main.go
go build -ldflags '-w -s' -o    hacash_miner_relay_service_2022_01_25_01  miner/run/minerrelayservice/main.go
go build -ldflags '-w -s' -o hacash_desktop_offline_wallet_2022_01_25_01  pcwallet/main/main.go
go build -ldflags '-w -s' -o              hacash_cmdwallet_2022_01_25_01  cmdwallet/run/main/main.go

go build -ldflags '-w -s' -o    hacash_channelpay_servicer_2022_01_25_01  channelpay/run/servicer/main.go
go build -ldflags '-w -s' -o      hacash_channelpay_client_2022_01_25_01  channelpay/run/client/main.go

cd ./x16rs/opencl && node pkgclfilego.js && cd ../../
go build -ldflags '-w -s' -o           hacash_miner_worker_2022_01_25_01  miner/run/minerworker/main.go

go build -ldflags '-w -s' -o                hacash_ranking_2022_12_09_01  github.com/hacash/service/ranking/


*/

const (
	DatabaseLowestVersion  int = 9  // Compatible database version number
	DatabaseCurrentVersion int = 12 // Current database version number
	//
	NodeVersionSuperMain    uint32 = 0            // Major version number
	NodeVersionSupport      uint32 = 1            // Compatible version number
	NodeVersionFeature      uint32 = 14           // Feature version number
	NodeVersionBuildCompile string = "20221127.1" // Build version number
	// Integrated version number system: 0.1.14 (20221127.1)
)

/**
 * start node
 */
func start() error {

	target_ini_file := "hacash.config.ini"
	//target_ini_file := "/home/shiqiujise/Desktop/Hacash/go/src/github.com/hacash/miner/run/main/test.ini"
	//target_ini_file := ""
	if len(os.Args) >= 2 {
		target_ini_file = os.Args[1]
	}

	target_ini_file = sys.AbsDir(target_ini_file)

	if target_ini_file != "" {
		fmt.Println("[Config] Load ini config file: \"" + target_ini_file + "\" at time:" + time.Now().Format("01/02 15:04:05"))
	}

	// Parse and load configuration file
	hinicnf, err := sys.LoadInicnf(target_ini_file)
	if err != nil {
		fmt.Println("[Config] ERROR TO LOAD CONFIG FILE: ", err.Error())
		return err
	}

	// Set database version
	hinicnf.SetDatabaseVersion(DatabaseCurrentVersion, DatabaseLowestVersion)

	// Judge whether the database version needs to be upgraded
	if hinicnf.Section("").Key("UseBlockChainV2").MustBool(false) {
		err = blockchain.CheckAndUpdateBlockchainDatabaseVersion(hinicnf)
	} else {
		err = blockchainv3.CheckAndUpdateBlockchainDatabaseVersion(hinicnf)
	}
	if err != nil {
		return err
	}

	//fmt.Println("=-===debugTestConfigSetHandle--------------")
	// debug test config set
	debugTestConfigSetHandle(hinicnf)

	//test_data_dir := "/home/shiqiujie/Desktop/Hacash/go/src/github.com/hacash/miner/run/minerpool/testdata"
	//hinicnf.SetMustDataDir(test_data_dir)

	isOpenMiner := hinicnf.Section("miner").Key("enable").MustBool(false)
	isOpenMinerServer := hinicnf.Section("minerserver").Key("enable").MustBool(false)
	isOpenMinerPool := hinicnf.Section("minerpool").Key("enable").MustBool(false)
	isOpenService := hinicnf.Section("service").Key("enable").MustBool(false)
	isOpenDiamondMiner := hinicnf.Section("diamondminer").Key("enable").MustBool(false)

	if (isOpenMinerServer || isOpenMinerPool) && !isOpenMiner {
		err := fmt.Errorf("[Error Exit] [Config] open [minerserver] or [minerpool] must open [miner] first.")
		return err
	}

	if isOpenDiamondMiner && isOpenMiner {
		err = fmt.Errorf("[Error Exit] [Config] Both [diamondminer] and [miner] cannot be turned on at the same time.")
		return err
	}

	// Check port occupancy
	p2pcnf := p2pv2.NewP2PConfig(hinicnf)
	p2port := p2pcnf.TCPListenPort
	portckconn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", p2port), time.Second)
	if err == nil {
		portckconn.Close() // Turn off port check
		return fmt.Errorf("Hacash P2P listen port %d already be occupied, is the node instance already started?", p2port)
	}

	// Formal start node
	hcnf := backend.NewBackendConfig(hinicnf)
	hnode, err := backend.NewBackend(hcnf)
	if err != nil {
		err = fmt.Errorf("backend.NewBackend Error: %s", err)
		return err
	}
	blockchainobj := hnode.BlockChain()

	txpool := memtxpool.NewMemTxPool(0, 1024*1024*50)
	txpool.SetBlockChain(blockchainobj)
	txpool.Start()

	// hnode set tx pool
	hnode.SetTxPool(txpool)

	// start
	err = hnode.Start()
	if err != nil {
		err = fmt.Errorf("backend.NewBackend.Start() Error: %s", err)
		return err
	}

	if isOpenMiner {

		mcnf := miner.NewMinerConfig(hinicnf)
		minernode := miner.NewMiner(mcnf)

		if isOpenMinerServer {

			// miner server
			mpcnf := minerserver.NewMinerConfig(hinicnf)
			miner_server := minerserver.NewMinerServer(mpcnf)

			err = miner_server.Start()
			if err != nil {
				err = fmt.Errorf("miner_server.Start() Error: %s", err)
				return err
			}

			// Set up POW server
			minernode.SetPowServer(miner_server)

			mpcnf1 := minerpool.NewMinerPoolConfig(hinicnf)
			miner_pool := minerpool.NewMinerPool(mpcnf1)
			miner_pool.SetBlockChain(blockchainobj)
			miner_pool.SetTxPool(txpool)

			// Set up POW server
			minernode.SetPowServer(miner_pool)

			// check reward address and password
			if !mcnf.Rewards.Equal(mpcnf1.RewardAccount.Address) {
				err = fmt.Errorf("[Config Error] miner rewards address must equal to miner pool rewards passward address.")
				fmt.Printf(err.Error())
				fmt.Printf("[配置错误] 矿池自动发送奖励的地址的密码应该是地址 %s 而不是地址 %s 的密码。\n", mcnf.Rewards.ToReadable(), mpcnf1.RewardAccount.AddressReadable)
				return err
			}

			err = miner_pool.Start()
			if err != nil {
				err = fmt.Errorf("miner_pool.Start() Error: %s", err)
				return err
			}

			cscnf := console.NewMinerConsoleConfig(hinicnf)
			console_service := console.NewMinerConsole(cscnf)
			console_service.SetMiningPool(miner_pool)

			err = console_service.Start() // http service
			if err != nil {
				err = fmt.Errorf("miner_server.Start() Error: %s", err)
				return err
			}

		} else {

			// full node local cpu
			lccnf := localcpu.NewFullNodePowWrapConfig(hinicnf)
			powwrap := localcpu.NewFullNodePowWrap(lccnf)

			// Set up POW server
			minernode.SetPowServer(powwrap)

		}

		// do mining
		minernode.SetBlockChain(blockchainobj)
		minernode.SetTxPool(txpool)
		err = minernode.Start()
		if err != nil {
			err = fmt.Errorf("minernode.Start() Error: %s", err)
			return err
		}

		err = minernode.StartMining()
		if err != nil {
			err = fmt.Errorf("minernode.StartMining() Error: %s", err)
			return err
		}

	} else {

		txpool.SetAutomaticallyCleanInvalidTransactions(true)

	}

	// http api service
	if isOpenService {

		// deprecated http api
		svcnf := deprecated.NewDeprecatedApiServiceConfig(hinicnf)
		if svcnf.HttpListenPort > 0 {
			deprecatedApi := deprecated.NewDeprecatedApiService(svcnf)
			deprecatedApi.SetBlockChain(blockchainobj)
			deprecatedApi.SetTxPool(txpool)
			deprecatedApi.SetBackend(hnode)
			err = deprecatedApi.Start()
			if err != nil {
				err = fmt.Errorf("deprecatedApi.Start() Error: %s", err)
				return err
			}
		}

		// rpc api
		rpccnf := rpc.NewRpcConfig(hinicnf)
		if rpccnf.HttpListenPort > 0 {
			rpcService := rpc.NewRpcService(rpccnf)
			rpcService.SetTxPool(txpool)
			rpcService.SetBackend(hnode)
			err = rpcService.Start()
			if err != nil {
				err = fmt.Errorf("rpcService.Start() Error: %s", err)
				return err
			}
		}
	}

	// diamond miner
	if isOpenDiamondMiner {

		dmcnf := diamondminer.NewDiamondMinerConfig(hinicnf)
		diamondMiner := diamondminer.NewDiamondMiner(dmcnf)
		diamondMiner.SetTxPool(txpool)
		diamondMiner.SetBlockChain(blockchainobj)

		err = diamondMiner.Start() // start do mining
		if err != nil {
			err = fmt.Errorf("diamondMiner.Start() Error: %s", err)
			return err
		}

	}

	//go func() {
	//	time.Sleep(time.Second * 3)
	//	Test_print_dmdname(hnode.BlockChain().State())
	//}()

	return nil
}

func main() {

	printAllVersion()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// start miner node
	err := start()
	if err != nil {
		fmt.Printf("\n-------- Hacash Node Run Failed Error: --------\n\n")
		fmt.Println(err.Error()) // print error
		fmt.Printf("\n-------- Hacash Node Run Failed end.   --------\n\n")
	}

	s := <-c
	fmt.Println("Got signal:", s)

}

////////////////////////////////

func printAllVersion() {

	// Print version number
	fmt.Printf("[Version] Hacash node software: %d.%d.%d(%s), ",
		NodeVersionSuperMain,
		NodeVersionSupport,
		NodeVersionFeature,
		NodeVersionBuildCompile)

	// p2p
	fmt.Printf("p2p compatible: block version[%d], transaction type [%d], action kind [%d], repair num [%d]\n",
		blocks.BlockVersion,
		blocks.TransactionType,
		blocks.ActionKind,
		blocks.RepairVersion)

}

/////////////////////////////////////////////////////

func debugTestConfigSetHandle(hinicnf *sys.Inicnf) {

	rootsec := hinicnf.Section("")

	// Global test mark testdebuglocaldevelopmentmark
	sys.TestDebugLocalDevelopmentMark = rootsec.Key("TestDebugLocalDevelopmentMark").MustBool(false)

	// test set start
	if adjustTargetDifficultyNumberOfBlocks := rootsec.Key("AdjustTargetDifficultyNumberOfBlocks").MustUint64(0); adjustTargetDifficultyNumberOfBlocks > 0 {
		mint.AdjustTargetDifficultyNumberOfBlocks = adjustTargetDifficultyNumberOfBlocks
	}
	if eachBlockRequiredTargetTime := rootsec.Key("EachBlockRequiredTargetTime").MustUint64(0); eachBlockRequiredTargetTime > 0 {
		mint.EachBlockRequiredTargetTime = eachBlockRequiredTargetTime
	}
	// test set end
}

/////////////////////////////////////////////////////

func Test_print_dmdname(state interfacev2.ChainState) {

	store := state.BlockStore()

	efcaddrs := ``

	adddrs := map[string]bool{}
	aaas := strings.Split(efcaddrs, "\n")
	for _, v := range aaas {
		if len(v) > 10 {
			adddrs[v] = true
		}
	}

	alladdrdmds := map[string][]string{}

	for i := uint32(1); i < 30000; i++ {
		dmd, e := store.ReadDiamondByNumber(i)
		if e != nil || dmd == nil {
			break
		}
		dia, _ := state.Diamond(dmd.Diamond)
		if dia == nil {
			break
		}
		addr := dia.Address.ToReadable()
		if _, o1 := adddrs[addr]; !o1 {
			continue
		}
		if list, ok := alladdrdmds[addr]; ok {
			alladdrdmds[addr] = append(list, string(dmd.Diamond))
		} else {
			alladdrdmds[addr] = []string{string(dmd.Diamond)}
		}
		fmt.Printf(",%d", i)
	}

	fmt.Println("\n\n\n\n ")
	// Print all
	for i, v := range alladdrdmds {
		fmt.Println(i + ":")
		for _, d := range v {
			fmt.Printf("%s,", d)
		}
		fmt.Println("\n ")
	}
	fmt.Println("\n\n\n\n ")

}
