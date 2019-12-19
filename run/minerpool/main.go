package main

import (
	"fmt"
	"github.com/hacash/core/sys"
	"github.com/hacash/miner/memtxpool"
	"github.com/hacash/miner/miner"
	"github.com/hacash/miner/minerpool"
	"github.com/hacash/node/backend"
	"os"
	"os/signal"
	"time"
)

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	test_ini := "/home/shiqiujie/Desktop/Hacash/go/src/github.com/hacash/miner/run/minerpool/test.ini"
	//test_ini := ""
	if len(os.Args) >= 2 {
		test_ini = os.Args[1]
	}

	if test_ini != "" {
		fmt.Println("Load ini config file: \"" + test_ini + "\" at time:" + time.Now().Format("01/02 15:04:05"))
	}

	hinicnf, _ := sys.LoadInicnf(test_ini)

	test_data_dir := "/home/shiqiujie/Desktop/Hacash/go/src/github.com/hacash/miner/run/minerpool/testdata"
	hinicnf.SetMustDataDir(test_data_dir)

	hcnf := backend.NewBackendConfig(hinicnf)
	hnode, err := backend.NewBackend(hcnf)
	if err != nil {
		fmt.Println(err)
		return
	}

	// start
	hnode.Start()

	txpool := memtxpool.NewMemTxPool(0, 1024*1024*50)
	txpool.SetBlockChain(hnode.GetBlockChain())

	//lccnf := localcpu.NewPowWrapConfig(hinicnf)
	//powwrap := localcpu.NewPowWrap(lccnf)

	mpcnf := minerpool.NewMinerPoolConfig(hinicnf)
	miner_pool := minerpool.NewMinerPool(mpcnf)
	miner_pool.SetBlockChain(hnode.GetBlockChain())
	miner_pool.Start()

	mcnf := miner.NewMinerConfig(hinicnf)
	miner := miner.NewMiner(mcnf)

	miner.SetBlockChain(hnode.GetBlockChain())
	miner.SetTxPool(txpool)
	miner.SetPowServer(miner_pool)

	miner.Start()

	// do mining
	miner.StartMining()

	// download block datas
	//hnode.DownloadBlocksDataFromWebSocketApi("ws://127.0.0.1:3338/websocket", 1)

	s := <-c
	fmt.Println("Got signal:", s)

}
