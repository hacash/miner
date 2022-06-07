package main

import (
	"fmt"
	"github.com/hacash/core/sys"
	"github.com/hacash/miner/minerrelayservice"
	"os"
	"os/signal"
	"time"
)

/**

go build -ldflags '-w -s' -o miner_relay_service_2021_04_02 github.com/hacash/miner/run/minerrelayservice


TEST:

go build -ldflags '-w -s' -o miner_relay_service_2021_04_02  miner/run/minerrelayservice/main.go && ./miner_relay_service_2021_04_02 ./miner/run/minerrelayservice/relayservice.config.ini


go build -o ./test/relayservice1  miner/run/minerrelayservice/main.go && ./test/relayservice1 ./rs1.ini

*/

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	target_ini_file := "relayservice.config.ini"
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

	// miner relay service
	cnf := minerrelayservice.NewMinerRelayServiceConfig(hinicnf)
	service := minerrelayservice.NewRelayService(cnf)

	// start-up
	service.Start()

	s := <-c
	fmt.Println("Got signal:", s)

}
