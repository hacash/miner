package message

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfacev2"
	"github.com/hacash/core/transactions"
)

/**

下发通知挖矿的区块

*/

type MsgPendingMiningBlockStuff struct {
	BlockHeadMeta                          interfacev2.Block                    // 区块的 head 和 meta
	CoinbaseTx                             *transactions.Transaction_0_Coinbase // coinbase 交易
	MrklRelatedTreeListForCoinbaseTxModify []fields.Hash                        // 默克尔树关联哈希
	// cache data
	mint_successed bool
}

// interfaces.PowWorkerMiningStuffItem

func (m *MsgPendingMiningBlockStuff) SetMiningSuccessed(ok bool) {
	m.mint_successed = ok
}
func (m MsgPendingMiningBlockStuff) GetMiningSuccessed() bool {
	return m.mint_successed
}
func (m MsgPendingMiningBlockStuff) GetHeadMetaBlock() interfacev2.Block {
	return m.BlockHeadMeta
}
func (m MsgPendingMiningBlockStuff) GetCoinbaseNonce() []byte {
	return m.CoinbaseTx.MinerNonce
}
func (m MsgPendingMiningBlockStuff) GetHeadNonce() []byte {
	ncbts := make([]byte, 4)
	binary.BigEndian.PutUint32(ncbts, m.BlockHeadMeta.GetNonce())
	return ncbts
}
func (m MsgPendingMiningBlockStuff) SetHeadNonce(nonce []byte) {
	m.BlockHeadMeta.SetNonce(binary.BigEndian.Uint32(nonce))
}
func (m MsgPendingMiningBlockStuff) CopyForMiningByRandomSetCoinbaseNonce() interfacev2.PowWorkerMiningStuffItem {
	newcbnonce := make([]byte, 32)
	rand.Read(newcbnonce)
	//fmt.Println(newcbnonce)
	newstuff, _ := m.CalculateBlockHashByBothNonce([]byte{0, 0, 0, 0}, newcbnonce, true) // copy
	//fmt.Println(newstuff.GetHeadMetaBlock().GetMrklRoot())
	return newstuff
}

// 创建 mining stuff
func CreatePendingMiningBlockStuffByBlock(block interfacev2.Block) (*MsgPendingMiningBlockStuff, error) {
	stuff := &MsgPendingMiningBlockStuff{
		BlockHeadMeta: block.CopyForMining(),
	}

	trxs := block.GetTransactions()
	if len(trxs) < 1 {
		return nil, fmt.Errorf("Block Transactions len error")
	}
	cbtrs := trxs[0]
	cbtx, ok := cbtrs.(*transactions.Transaction_0_Coinbase)
	if ok == false {
		return nil, fmt.Errorf("Block Transaction_0_Coinbase error")
	}
	stuff.CoinbaseTx = cbtx
	stuff.MrklRelatedTreeListForCoinbaseTxModify = blocks.PickMrklListForCoinbaseTxModify(trxs)
	return stuff, nil
}

// 序列化
func (m MsgPendingMiningBlockStuff) Serialize() []byte {
	buf := bytes.NewBuffer([]byte{})
	b1, _ := m.BlockHeadMeta.SerializeExcludeTransactions()
	buf.Write(b1)
	b2, _ := m.CoinbaseTx.Serialize()
	buf.Write(b2)
	mrklsize := len(m.MrklRelatedTreeListForCoinbaseTxModify)
	mrklsizebytes := []byte{0, 0}
	binary.BigEndian.PutUint16(mrklsizebytes, uint16(mrklsize))
	buf.Write(mrklsizebytes)
	for i := 0; i < int(mrklsize); i++ {
		buf.Write(m.MrklRelatedTreeListForCoinbaseTxModify[i])
	}
	// all ok
	return buf.Bytes()
}

// 反序列化
func (m *MsgPendingMiningBlockStuff) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error = nil
	m.BlockHeadMeta, seek, e = blocks.ParseExcludeTransactions(buf, seek)
	if e != nil {
		return 0, e
	}
	var trs interfacev2.Transaction = nil
	trs, seek, e = transactions.ParseTransaction(buf, seek)
	if e != nil {
		return 0, e
	}
	cbtx, ok := trs.(*transactions.Transaction_0_Coinbase)
	if ok == false {
		return 0, fmt.Errorf("tx must be Transaction_0_Coinbase")
	}
	m.CoinbaseTx = cbtx
	// mrkl
	if len(buf) < int(seek)+2 {
		return 0, fmt.Errorf("buf len error")
	}
	mrklsize := binary.BigEndian.Uint16(buf[seek : seek+2])
	hxlist := make([]fields.Hash, 0)
	seek += 2
	for i := 0; i < int(mrklsize); i++ {
		if len(buf) < int(seek)+32 {
			return 0, fmt.Errorf("buf len error")
		}
		hxlist = append(hxlist, buf[seek:seek+32])
		seek += 32
	}
	m.MrklRelatedTreeListForCoinbaseTxModify = hxlist
	// all ok
	return seek, nil
}

// 通过设置nonce值计算 区块哈希
func (m MsgPendingMiningBlockStuff) CalculateBlockHashByBothNonce(headNonce fields.Bytes4, coinbaseNonce fields.Bytes32, retcopy bool) (*MsgPendingMiningBlockStuff, fields.Hash) {
	//
	newblock := m.BlockHeadMeta.CopyForMining()
	newblock.SetNonce(binary.BigEndian.Uint32(headNonce))
	/// copy coinbase hash
	cbnonce := make([]byte, 32)
	copy(cbnonce, coinbaseNonce)
	coinbasetxcopy := m.CoinbaseTx.Copy()
	newcbtx := coinbasetxcopy.(*transactions.Transaction_0_Coinbase)
	if newcbtx == nil {
		panic("m.CoinbaseTx must be a *transactions.Transaction_0_Coinbase")
	}
	newcbtx.MinerNonce = cbnonce
	// 计算 mrkl root
	cbtxhx := newcbtx.Hash()
	mrklroot := blocks.CalculateMrklRootByCoinbaseTxModify(cbtxhx, m.MrklRelatedTreeListForCoinbaseTxModify)
	newblock.SetMrklRoot(mrklroot)
	// hash
	blkhx := newblock.HashFresh()
	// copy
	var copystuff *MsgPendingMiningBlockStuff = nil
	if retcopy {
		copystuff = &MsgPendingMiningBlockStuff{
			BlockHeadMeta:                          newblock,
			CoinbaseTx:                             newcbtx,
			MrklRelatedTreeListForCoinbaseTxModify: m.MrklRelatedTreeListForCoinbaseTxModify,
		}
		realtxs := newblock.GetTransactions()
		if len(realtxs) > 0 {
			realtxs[0] = newcbtx // copy coinbase tx
		}
	}
	return copystuff, blkhx
}

// 通过设置nonce值计算 区块哈希
func (m MsgPendingMiningBlockStuff) CalculateBlockHashByMiningResult(result *MsgReportMiningResult, retcopy bool) (*MsgPendingMiningBlockStuff, fields.Hash) {
	return m.CalculateBlockHashByBothNonce(result.HeadNonce, result.CoinbaseNonce, retcopy)
}
