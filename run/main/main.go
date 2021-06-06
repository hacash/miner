package main

import (
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
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
	"github.com/hacash/node/backend"
	deprecated "github.com/hacash/service/deprecated"
	rpc "github.com/hacash/service/rpc"
	"os"
	"os/signal"
	"strings"
	"time"
)

/**

export GOPATH=/media/yangjie/500GB/hacash/go

go build -o test/mainnet  miner/run/main/main.go && ./test/mainnet  mainnet.ini
go build -o test/test1    miner/run/main/main.go && ./test/test1    test1.ini
go build -o test/test2    miner/run/main/main.go && ./test/test2    test2.ini
go build -o test/test3    miner/run/main/main.go && ./test/test3    test3.ini
go build -o test/pcwallet pcwallet/main/main.go  && ./test/pcwallet

*/

/**

编译发布版本：

go build -ldflags '-w -s' -o                   hacash_node_2021_05_15_01  miner/run/main/main.go
go build -ldflags '-w -s' -o      hacash_miner_pool_worker_2021_05_12_01  miner/run/minerpoolworker/main.go
go build -ldflags '-w -s' -o              hacash_cmdwallet_2021_05_12_01  cmdwallet/run/main/main.go

cd ./x16rs/opencl && node pkgclfilego.js && cd ../../
go build -ldflags '-w -s' -o           hacash_miner_worker_2021_05_12_01  miner/run/minerworker/main.go
go build -ldflags '-w -s' -o    hacash_miner_relay_service_2021_05_12_01 miner/run/minerrelayservice/main.go
go build -ldflags '-w -s' -o hacash_desktop_offline_wallet_2021_05_18_01  pcwallet/main/main.go


*/

const (
	NodeVersionSuperMain    uint32 = 0            // 主版本号
	NodeVersionSupport      uint32 = 1            // 兼容版本号
	NodeVersionFeature      uint32 = 4            // 特征版本号
	NodeVersionBuildCompile string = "20210515.1" // 编译版本号
	// 结合成综合版本号体系：   0.1.4(20210515.1)
)

func main() {

	printAllVersion()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

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

	// 解析并载入配置文件
	hinicnf, err := sys.LoadInicnf(target_ini_file)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	// 判断数据库版本是否需要升级
	blockchain.CheckAndUpdateBlockchainDatabaseVersion(hinicnf)

	//fmt.Println("=-===debugTestConfigSetHandle--------------")
	// debug test config set
	debugTestConfigSetHandle(hinicnf)

	//test_data_dir := "/home/shiqiujie/Desktop/Hacash/go/src/github.com/hacash/miner/run/minerpool/testdata"
	//hinicnf.SetMustDataDir(test_data_dir)

	hcnf := backend.NewBackendConfig(hinicnf)
	hnode, err := backend.NewBackend(hcnf)
	if err != nil {
		fmt.Println("backend.NewBackend Error", err)
		return
	}
	blockchainobj := hnode.BlockChain()

	txpool := memtxpool.NewMemTxPool(0, 1024*1024*50)
	txpool.SetBlockChain(blockchainobj)
	txpool.Start()

	// hnode set tx pool
	hnode.SetTxPool(txpool)

	// start
	hnode.Start()

	isOpenMiner := hinicnf.Section("miner").Key("enable").MustBool(false)
	isOpenMinerServer := hinicnf.Section("minerserver").Key("enable").MustBool(false)
	isOpenMinerPool := hinicnf.Section("minerpool").Key("enable").MustBool(false)
	isOpenService := hinicnf.Section("service").Key("enable").MustBool(false)
	isOpenDiamondMiner := hinicnf.Section("diamondminer").Key("enable").MustBool(false)

	if (isOpenMinerServer || isOpenMinerPool) && !isOpenMiner {
		fmt.Println("[Error Exit] [Config] open [minerserver] or [minerpool] must open [miner] first.")
		os.Exit(0)
	}

	if isOpenMiner {

		mcnf := miner.NewMinerConfig(hinicnf)
		miner := miner.NewMiner(mcnf)

		if isOpenMinerServer {

			// miner server
			mpcnf := minerserver.NewMinerConfig(hinicnf)
			miner_server := minerserver.NewMinerServer(mpcnf)

			miner_server.Start()

			// 设置 pow server
			miner.SetPowServer(miner_server)

		} else if isOpenMinerPool {

			mpcnf := minerpool.NewMinerPoolConfig(hinicnf)
			miner_pool := minerpool.NewMinerPool(mpcnf)
			miner_pool.SetBlockChain(blockchainobj)
			miner_pool.SetTxPool(txpool)

			// 设置 pow server
			miner.SetPowServer(miner_pool)

			// check reward address and password
			if !mcnf.Rewards.Equal(mpcnf.RewardAccount.Address) {
				fmt.Println("[Config Error] miner rewards address must equal to miner pool rewards passward address.")
				fmt.Printf("[配置错误] 矿池自动发送奖励的地址的密码应该是地址 %s 而不是地址 %s 的密码。\n", mcnf.Rewards.ToReadable(), mpcnf.RewardAccount.AddressReadable)
				os.Exit(0)
			}

			miner_pool.Start()

			cscnf := console.NewMinerConsoleConfig(hinicnf)
			console_service := console.NewMinerConsole(cscnf)
			console_service.SetMiningPool(miner_pool)

			console_service.Start() // http service

		} else {

			// full node local cpu
			lccnf := localcpu.NewFullNodePowWrapConfig(hinicnf)
			powwrap := localcpu.NewFullNodePowWrap(lccnf)

			// 设置 pow server
			miner.SetPowServer(powwrap)

		}

		// do mining
		miner.SetBlockChain(blockchainobj)
		miner.SetTxPool(txpool)
		miner.Start()
		miner.StartMining()

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
			deprecatedApi.Start()
		}

		// rpc api
		rpccnf := rpc.NewRpcConfig(hinicnf)
		if rpccnf.HttpListenPort > 0 {
			rpcService := rpc.NewRpcService(rpccnf)
			rpcService.SetTxPool(txpool)
			rpcService.SetBackend(hnode)
			rpcService.Start()
		}
	}

	// diamond miner
	if isOpenDiamondMiner {

		dmcnf := diamondminer.NewDiamondMinerConfig(hinicnf)
		diamondMiner := diamondminer.NewDiamondMiner(dmcnf)
		diamondMiner.SetTxPool(txpool)
		diamondMiner.SetBlockChain(blockchainobj)

		diamondMiner.Start() // start do mining

	}

	// download block datas
	wsaddr := hinicnf.Section("").Key("first_download_block_datas_websocket_addr").MustString("")
	wsurl1 := "ws://" + wsaddr + "/ws/download"
	if wsaddr != "" {
		//time.Sleep( time.Second * 3 )
		hnode.DownloadBlocksDataFromWebSocketApi(wsurl1, 1)
	}

	// sync block
	syncblockwsaddr := hinicnf.Section("").Key("sync_block_websocket_addr").MustString("")
	syncblocktimesleep := hinicnf.Section("").Key("sync_block_websocket_timesleep").MustUint(60 * 3)
	wssyncurl := "ws://" + syncblockwsaddr + "/ws/sync"
	if syncblockwsaddr != "" {
		fmt.Println("Sync new block from", wssyncurl)
		go func() {
			for {
				//time.Sleep(time.Minute * 3)
				time.Sleep(time.Second * time.Duration(syncblocktimesleep))
				err := hnode.SyncBlockFromWebSocketApi(wssyncurl)
				if err != nil {
					fmt.Println("SyncBlockFromWebSocketApi Error:", err.Error())
				}
			}
		}()
	}

	//go func() {
	//	time.Sleep(time.Second * 3)
	//	Test_print_dmdname(hnode.BlockChain().State())
	//}()

	s := <-c
	fmt.Println("Got signal:", s)

}

////////////////////////////////

func printAllVersion() {

	// 打印版本号
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

	// 全局测试标记 TestDebugLocalDevelopmentMark
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

func Test_print_dmdname(state interfaces.ChainState) {

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
		dia := state.Diamond(dmd.Diamond)
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
	// 打印全部
	for i, v := range alladdrdmds {
		fmt.Println(i + ":")
		for _, d := range v {
			fmt.Printf("%s,", d)
		}
		fmt.Println("\n ")
	}
	fmt.Println("\n\n\n\n ")

}
