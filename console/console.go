package console

import (
	"fmt"
	"github.com/hacash/miner/minerpool"
	"os"
)

type MinerConsole struct {
	config *MinerConsoleConfig

	pool *minerpool.MinerPool
}

func NewMinerConsole(cnf *MinerConsoleConfig) *MinerConsole {

	cons := &MinerConsole{
		config: cnf,
		pool:   nil,
	}

	return cons
}

func (mc *MinerConsole) Start() {
	if mc.pool == nil {
		panic(fmt.Errorf("miner pool not be set."))
	}

	err := mc.startHttpService()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

}

func (mc *MinerConsole) SetMiningPool(pool *minerpool.MinerPool) {
	mc.pool = pool
}
