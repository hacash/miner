package localcpu

import (
	"encoding/binary"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/sys"
	"github.com/hacash/miner/message"
	"github.com/hacash/mint/coinbase"
)

type FullNodePowWrapConfig struct {
	cnffile *sys.Inicnf
}

func NewEmptyFullNodePowWrapConfig() *FullNodePowWrapConfig {
	cnf := &FullNodePowWrapConfig{}
	return cnf
}

//////////////////////////////////////////////////

func NewFullNodePowWrapConfig(cnffile *sys.Inicnf) *FullNodePowWrapConfig {
	cnf := NewEmptyFullNodePowWrapConfig()
	cnf.cnffile = cnffile
	return cnf

}

///////////////////////////////////////////////

type FullNodePowWrap struct {
	config *FullNodePowWrapConfig

	master *LocalCPUPowMaster
}

func NewFullNodePowWrap(cnf *FullNodePowWrapConfig) *FullNodePowWrap {

	wrap := &FullNodePowWrap{
		config: cnf,
	}

	lccnf := NewLocalCPUPowMasterConfig(cnf.cnffile)

	cnfsection := cnf.cnffile.Section("miner")
	// supervene
	lccnf.Concurrent = uint32(cnfsection.Key("supervene").MustUint(1))

	powmaster := NewLocalCPUPowMaster(lccnf)

	wrap.master = powmaster

	return wrap
}

//////////////////////////////////////////////////////////////////

func (p *FullNodePowWrap) StopMining() {
	p.master.StopMining()
}

func (p *FullNodePowWrap) Excavate(inputBlock interfaces.Block, outputBlockCh chan interfaces.Block) {

	var coinbaseMsgNum uint32 = 0

	for {
		if coinbaseMsgNum > 0 {
			coinbase.UpdateBlockCoinbaseMessageForMiner(inputBlock, coinbaseMsgNum)
			inputBlock.SetMrklRoot(blocks.CalculateMrklRoot(inputBlock.GetTransactions())) // update mrkl
		}
		p.master.SetCoinbaseMsgNum(coinbaseMsgNum)
		var outputCh = make(chan message.PowMasterMsg, 1)
		p.master.Excavate(inputBlock, outputCh)
		output := <-outputCh
		if output.Status == message.PowMasterMsgStatusContinue {
			// continue next
			coinbaseMsgNum++
			//fmt.Println( "output.Status == Continue  coinbaseMsgNum ++ ", coinbaseMsgNum)
			continue
		}
		if output.Status == message.PowMasterMsgStatusStop {
			return // do nothing
		}
		if output.Status == message.PowMasterMsgStatusSuccess {
			output.BlockHeadMeta.SetNonce(binary.BigEndian.Uint32(output.NonceBytes))
			output.BlockHeadMeta.Fresh()
			outputBlockCh <- output.BlockHeadMeta
			return // success
		}
		fmt.Println("[Mining Error]", output.Status)
	}

}
