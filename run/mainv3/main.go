package main

import (
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/sys"
	"github.com/hacash/mint"
	"github.com/hacash/mint/blockchainv3"
	"math/rand"
	"os"
	"os/signal"
	"time"
)

const (
	// 数据库版本号
	DatabaseLowestVersion  int = 10 // 兼容版本号
	DatabaseCurrentVersion int = 10 // 版本号
	// 软件版本号
	NodeVersionSuperMain    uint32 = 0            // 主版本号
	NodeVersionSupport      uint32 = 1            // 兼容版本号
	NodeVersionFeature      uint32 = 8            // 特征版本号
	NodeVersionBuildCompile string = "20211127.1" // 编译版本号
	// 结合成综合版本号体系：   0.1.8(20211127.1)
)

func main() {

	printAllVersion()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	rand.Seed(time.Now().Unix())

	// start miner node
	err := start()
	if err != nil {
		fmt.Println("\n-------- Hacash Node Run Failed Error: --------\n")
		fmt.Println(err.Error()) // print error
		fmt.Println("\n-------- Hacash Node Run Failed end.   --------\n")
	}

	s := <-c
	fmt.Println("Got signal:", s)

}

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

	// 解析并载入配置文件
	hinicnf, err := sys.LoadInicnf(target_ini_file)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	// 设置数据库版本
	hinicnf.SetDatabaseVersion(DatabaseCurrentVersion, DatabaseLowestVersion)

	// 创建区块链实例
	bccnf := blockchainv3.NewBlockChainConfig(hinicnf)
	blockChainObj, e := blockchainv3.NewBlockChain(bccnf)
	if e != nil {
		return e
	}

	fmt.Println(blockChainObj)
	return nil
}

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
