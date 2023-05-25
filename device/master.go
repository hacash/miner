package device

import (
	"bytes"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/transactions"
	itfcs "github.com/hacash/miner/interfaces"
	"github.com/hacash/mint/difficulty"
)

type PoWMasterMng struct {
	worker itfcs.PoWWorker
}

func NewPoWMasterMng(alloter itfcs.PoWExecute) *PoWMasterMng {
	var worker = NewPoWWorkerMng(alloter)
	return &PoWMasterMng{
		worker: worker,
	}
}

func (m *PoWMasterMng) Init() error {
	return m.worker.Init()
}

func (m *PoWMasterMng) StopMining() {
	m.worker.StopMining()
}

// find a block
func (m *PoWMasterMng) DoMining(input interfaces.Block, resCh chan interfaces.Block) error {
	// stop prev mining
	m.worker.StopMining()
	// prepare stuff
	trslist := input.GetTrsList()
	if len(trslist) < 1 {
		return fmt.Errorf("tx len cannot less than 1")
	}
	coinbase_tx, ok := trslist[0].(*transactions.Transaction_0_Coinbase)
	if !ok {
		return fmt.Errorf("first tx must is coinbase tx")
	}
	mkrltree := blocks.PickMrklListForCoinbaseTxModify(trslist)
	// ok
	var stuff = &itfcs.PoWStuffOverallData{
		BlockHeadMeta:     input.CopyHeadMetaForMining(),
		CoinbaseTx:        *coinbase_tx.CopyForMining(),
		MrklCheckTreeList: mkrltree,
	}
	// do mining
	go func(mnstuff *itfcs.PoWStuffOverallData,
		resCh chan interfaces.Block, useblock interfaces.Block,
		coinbase_tx *transactions.Transaction_0_Coinbase) {
		var result, e = m.worker.DoMining(mnstuff)
		if e != nil {
			return
		}
		if result == nil {
			return
		}
		if !result.FindSuccess.Check() {
			return
		}
		// check find
		useblock.SetNonce(uint32(result.BlockNonce))
		coinbase_tx.MinerNonce = result.CoinbaseNonce
		mkrlroot := blocks.CalculateMrklRootByCoinbaseTxModify(coinbase_tx.Hash(), stuff.MrklCheckTreeList)
		useblock.SetMrklRoot(mkrlroot)
		res_block_hash := useblock.HashFresh()
		target_diff_hash := difficulty.DifficultyUint32ToHashForAntimatter(useblock.GetDifficulty())
		if bytes.Compare(res_block_hash, target_diff_hash) == 1 {
			return // check fail
		}
		// SUCCESS find a block !
		resCh <- useblock
	}(stuff, resCh, input, coinbase_tx)

	return nil
}
