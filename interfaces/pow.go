package interfaces

import (
	"bytes"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/transactions"
)

type PoWStuffBriefData struct {
	BlockHeadMeta interfaces.Block
	CoinbaseNonce fields.Hash
}

type PoWStuffOverallData struct {
	BlockHeadMeta     interfaces.Block
	CoinbaseTx        transactions.Transaction_0_Coinbase
	MrklCheckTreeList []fields.Hash
}

func (p *PoWStuffOverallData) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	b1, _ := p.BlockHeadMeta.SerializeExcludeTransactions()
	b2, _ := p.CoinbaseTx.Serialize()
	tree := fields.CreateHashListMax65535(p.MrklCheckTreeList)
	b3, _ := tree.Serialize()
	buf.Write(b1)
	buf.Write(b2)
	buf.Write(b3)
	return buf.Bytes(), nil
}

func (p *PoWStuffOverallData) Parse(buf []byte, seek uint32) (uint32, error) {
	//var e error = nil
	p.BlockHeadMeta, seek, _ = blocks.ParseExcludeTransactions(buf, seek)

	var trs interfaces.Transaction
	trs, seek, _ = transactions.ParseTransaction(buf, seek)
	p.CoinbaseTx = *trs.(*transactions.Transaction_0_Coinbase)

	tree := fields.CreateHashListMax65535(nil)
	seek, _ = tree.Parse(buf, seek)
	p.MrklCheckTreeList = tree.Hashs
	return seek, nil
}

func (p *PoWStuffOverallData) CalculateBlockHashByMiningResult(result *PoWResultShortData, if_block bool) (fields.Hash, interfaces.Block, error) {
	var curhei = p.BlockHeadMeta.GetHeight()
	var reshei = uint64(result.BlockHeight)
	if reshei != curhei {
		return nil, nil, fmt.Errorf("block height error: need %d but got %d", curhei, reshei)
	}
	// replace
	cbtx := p.CoinbaseTx.CopyForMining()
	cbtx.MinerNonce = result.CoinbaseNonce
	var cb_hash = cbtx.Hash()
	// mklrRoot
	var useblk interfaces.Block
	if if_block {
		useblk = p.BlockHeadMeta.CopyForMining()
		trslist := useblk.GetTrsList()
		trslist[0] = cbtx
		useblk.SetTrsList(trslist)
	} else {
		useblk = p.BlockHeadMeta.CopyHeadMetaForMining()
	}
	mklr_root := blocks.CalculateMrklRootByCoinbaseTxModify(cb_hash, p.MrklCheckTreeList)
	useblk.SetMrklRoot(mklr_root)
	useblk.SetNonce(uint32(result.BlockNonce))
	useblk.Fresh()
	// ok end
	return useblk.HashFresh(), useblk, nil
}

type PoWResultShortData struct {
	FindSuccess   fields.Bool
	BlockHeight   fields.BlockHeight
	BlockNonce    fields.VarUint4
	CoinbaseNonce fields.Hash
}

func (p *PoWResultShortData) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	b0, _ := p.FindSuccess.Serialize()
	b1, _ := p.BlockHeight.Serialize()
	b2, _ := p.BlockNonce.Serialize()
	b3, _ := p.CoinbaseNonce.Serialize()
	buf.Write(b0)
	buf.Write(b1)
	buf.Write(b2)
	buf.Write(b3)
	return buf.Bytes(), nil
}

func (p *PoWResultShortData) Parse(buf []byte, seek uint32) (uint32, error) {
	//var e error = nil
	seek, _ = p.FindSuccess.Parse(buf, seek)
	seek, _ = p.BlockHeight.Parse(buf, seek)
	seek, _ = p.BlockNonce.Parse(buf, seek)
	seek, _ = p.CoinbaseNonce.Parse(buf, seek)
	return seek, nil
}

type PoWResultData struct {
	PoWResultShortData
	ResultHash fields.Hash
}

func (p *PoWResultData) GetShortData() *PoWResultShortData {
	return &PoWResultShortData{
		p.FindSuccess,
		p.BlockHeight,
		p.BlockNonce,
		p.CoinbaseNonce,
	}
}

////////////////////////////////

type PoWConfig interface {
	IsDetailLog() bool
}

type PoWMaster interface {
	Config() PoWConfig
	Init() error
	DoMining(input interfaces.Block, resCh chan interfaces.Block) error // find a block
	StopMining()                                                        // stop all
}

type PoWWorker interface {
	Config() PoWConfig
	CloseUploadHashrate()
	Init() error
	DoMining(input *PoWStuffOverallData) (*PoWResultData, error) // find a block
	StopMining()                                                 // stop all
}

type PoWDevice interface {
	Config() PoWConfig
	Init() error
	DoMining(stopmark *byte, inputCh chan *PoWStuffBriefData) (*PoWResultData, error) // find a block
	StopMining()                                                                      // stop all
}

type PoWThread interface {
	Config() PoWConfig
	Init() error
	DoMining(stopmark *byte, target_hash fields.Hash, input PoWStuffBriefData, resCh chan *PoWResultData) error // find a block
	StopMining()                                                                                                // stop all
}

type PoWExecute interface {
	Config() PoWConfig
	Init() error
	DoMining(stopmark *byte, successmark *byte, input interfaces.Block, nonce_offset uint32) (*PoWResultData, error) // find a block

	GetNonceSpan() uint32
	ReportSpanTime(float64) // second
	//
	Allocate() chan PoWExecute
}
