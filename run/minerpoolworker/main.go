package main

import (
	"fmt"
	"github.com/hacash/core/sys"
	"github.com/hacash/miner/minerpoolworker"
	"os"
	"os/signal"
	"time"
)

/**

go build -o poolworker github.com/hacash/miner/run/minerpoolworker
go build -ldflags '-w -s' -o poolworker_2020_6_24 github.com/hacash/miner/run/minerpoolworker

*/

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	target_ini_file := "poolworker.config.ini"
	// target_ini_file := "/home/shiqiujie/Desktop/Hacash/go/src/github.com/hacash/miner/run/minerworker/test.ini"
	// target_ini_file := ""
	if len(os.Args) >= 2 {
		target_ini_file = os.Args[1]
	}

	target_ini_file = sys.AbsDir(target_ini_file)

	if target_ini_file != "" {
		fmt.Println("Load ini config file: \"" + target_ini_file + "\" at time:" + time.Now().Format("01/02 15:04:05"))
	}

	hinicnf, _ := sys.LoadInicnf(target_ini_file)

	cnf := minerpoolworker.NewMinerWorkerConfig(hinicnf)
	worker := minerpoolworker.NewMinerWorker(cnf)

	worker.Start()

	s := <-c
	fmt.Println("Got signal:", s)

}
