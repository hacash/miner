package main

import (
	"fmt"
	"github.com/hacash/core/sys"
	"github.com/hacash/miner/console"
	"github.com/hacash/miner/diamondminer"
	"github.com/hacash/miner/localcpu"
	"github.com/hacash/miner/memtxpool"
	"github.com/hacash/miner/miner"
	"github.com/hacash/miner/minerpool"
	"github.com/hacash/mint"
	"github.com/hacash/node/backend"
	rpc "github.com/hacash/service/deprecated"
	"os"
	"os/signal"
	"time"
)


/**

go build -o test/test1 miner/run/main/main.go && ./test/test1 test1.ini
go build -ldflags '-w -s' -o hacash_node_2020_01_09_3 miner/run/main/main.go

 */




func main() {

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
		fmt.Println("Load ini config file: \"" + target_ini_file + "\" at time:" + time.Now().Format("01/02 15:04:05"))
	}

	hinicnf, err := sys.LoadInicnf(target_ini_file)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	// debug test config set
	debugTestConfigSetHandle(hinicnf)

	//test_data_dir := "/home/shiqiujie/Desktop/Hacash/go/src/github.com/hacash/miner/run/minerpool/testdata"
	//hinicnf.SetMustDataDir(test_data_dir)

	hcnf := backend.NewBackendConfig(hinicnf)
	hnode, err := backend.NewBackend(hcnf)
	if err != nil {
		fmt.Println(err)
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

	isOpenMiner := hinicnf.Section("miner").Key("enable").MustString("false") == "true"
	isOpenMinerPool := hinicnf.Section("minerpool").Key("enable").MustString("false") == "true"
	isOpenService := hinicnf.Section("service").Key("enable").MustString("false") == "true"
	isOpenDiamondMiner := hinicnf.Section("diamondminer").Key("enable").MustString("false") == "true"

	if isOpenMiner {

		mcnf := miner.NewMinerConfig(hinicnf)
		miner := miner.NewMiner(mcnf)

		if isOpenMinerPool {

			mpcnf := minerpool.NewMinerPoolConfig(hinicnf)
			miner_pool := minerpool.NewMinerPool(mpcnf)
			miner_pool.SetBlockChain(blockchainobj)
			miner_pool.SetTxPool(txpool)

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

			// local cpu
			lccnf := localcpu.NewPowWrapConfig(hinicnf)
			powwrap := localcpu.NewPowWrap(lccnf)
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
		svcnf := rpc.NewDeprecatedApiServiceConfig(hinicnf)
		deprecatedApi := rpc.NewDeprecatedApiService(svcnf)
		deprecatedApi.SetBlockChain(blockchainobj)
		deprecatedApi.SetTxPool(txpool)
		deprecatedApi.SetBackend(hnode)
		deprecatedApi.Start()
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
	wsurl1 := "ws://"+wsaddr+"/ws/download"
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
				err := hnode.SyncBlockFromWebSocketApi( wssyncurl )
				if err != nil {
					fmt.Println("SyncBlockFromWebSocketApi Error:", err.Error())
				}
			}
		}()
	}

	s := <-c
	fmt.Println("Got signal:", s)

}

/////////////////////////////////////////////////////

func debugTestConfigSetHandle(hinicnf *sys.Inicnf) {

	// test set start
	if adjustTargetDifficultyNumberOfBlocks := hinicnf.Section("").Key("AdjustTargetDifficultyNumberOfBlocks").MustUint64(0); adjustTargetDifficultyNumberOfBlocks > 0 {
		mint.AdjustTargetDifficultyNumberOfBlocks = adjustTargetDifficultyNumberOfBlocks
	}
	if eachBlockRequiredTargetTime := hinicnf.Section("").Key("EachBlockRequiredTargetTime").MustUint64(0); eachBlockRequiredTargetTime > 0 {
		mint.EachBlockRequiredTargetTime = eachBlockRequiredTargetTime
	}
	// test set end
}