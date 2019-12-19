package main

import (
	"fmt"
	"github.com/hacash/core/sys"
	"github.com/hacash/miner/minerworker"
	"os"
	"os/signal"
	"time"
)

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	test_ini := "/home/shiqiujie/Desktop/Hacash/go/src/github.com/hacash/miner/run/minerworker/test.ini"
	//test_ini := ""
	if len(os.Args) >= 2 {
		test_ini = os.Args[1]
	}

	if test_ini != "" {
		fmt.Println("Load ini config file: \"" + test_ini + "\" at time:" + time.Now().Format("01/02 15:04:05"))
	}

	hinicnf, _ := sys.LoadInicnf(test_ini)

	cnf := minerworker.NewMinerWorkerConfig(hinicnf)
	worker := minerworker.NewMinerWorker(cnf)

	worker.Start()

	s := <-c
	fmt.Println("Got signal:", s)

}
