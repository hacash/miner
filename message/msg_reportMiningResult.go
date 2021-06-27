package message

import (
	"bytes"
	"github.com/hacash/core/fields"
)

/**

向上通报挖矿的结果

*/

type MsgReportMiningResult struct {
	MintSuccessed fields.Bool        // 挖掘成功 或者 报告算力
	BlockHeight   fields.BlockHeight // 挖掘的区块高度
	HeadNonce     fields.Bytes4      // block head nonce
	CoinbaseNonce fields.Bytes32     // coinbase nonce
}

// 序列化
func (m MsgReportMiningResult) Serialize() []byte {

	buf := bytes.NewBuffer([]byte{})
	b1, _ := m.MintSuccessed.Serialize()
	bhbts, _ := m.BlockHeight.Serialize()
	buf.Write(b1)
	buf.Write(bhbts)
	buf.Write(m.HeadNonce)
	buf.Write(m.CoinbaseNonce)

	return buf.Bytes()
}

// 反序列化
func (m *MsgReportMiningResult) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error = nil
	seek, e = m.MintSuccessed.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.BlockHeight.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.HeadNonce.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.CoinbaseNonce.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}
