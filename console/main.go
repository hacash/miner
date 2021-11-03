package console

import (
	"fmt"
	"github.com/hacash/miner/minerpool"
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

func (mc *MinerConsole) Start() error {
	if mc.pool == nil {
		return fmt.Errorf("miner pool not be set.")
	}

	err := mc.startHttpService()
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil

}

func (mc *MinerConsole) SetMiningPool(pool *minerpool.MinerPool) {
	mc.pool = pool
}
