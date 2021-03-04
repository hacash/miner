package message

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/transactions"
)

type MsgMiningBlockStuff struct {
	BlockHeadMeta                          interfaces.Block                     // 区块的 head 和 meta
	CoinbaseTx                             *transactions.Transaction_0_Coinbase // coinbase 交易
	MrklRelatedTreeListForCoinbaseTxModify []fields.Hash                        // 默克尔树关联哈希
}

// 序列化
func (m *MsgMiningBlockStuff) Serialize() []byte {
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
func (m *MsgMiningBlockStuff) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error = nil
	m.BlockHeadMeta, seek, e = blocks.ParseExcludeTransactions(buf, seek)
	if e != nil {
		return 0, e
	}
	var trs interfaces.Transaction = nil
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
func (m *MsgMiningBlockStuff) CalculateBlockHashBySetBothNonce(headNonce fields.Bytes4, coinbaseNonce fields.Bytes32) fields.Hash {

	newblock := m.BlockHeadMeta.CopyHeadMetaForMining()
	newblock.SetNonce(binary.BigEndian.Uint32(headNonce))
	// 计算 mrkl root
	cbtxhx := m.CoinbaseTx.Hash()
	mrklroot := blocks.CalculateMrklRootByCoinbaseTxModify(cbtxhx, m.MrklRelatedTreeListForCoinbaseTxModify)
	newblock.SetMrklRoot(mrklroot)
	// hash
	blkhx := newblock.HashFresh()
	return blkhx
}
