package message

import (
	"bytes"
	"github.com/hacash/core/fields"
)

/**

向上通报挖矿的结果

*/

type MsgReportMiningResult struct {
	MintSuccessed fields.Bool        // Mining success or reporting computing power
	BlockHeight   fields.BlockHeight // Excavated block height
	HeadNonce     fields.Bytes4      // block head nonce
	CoinbaseNonce fields.Bytes32     // coinbase nonce
}

// serialize
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

// Deserialization
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
