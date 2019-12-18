package message

import (
	"bytes"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
)

const (
	PowMasterMsgSize = blocks.BlockHeadSize + blocks.BlockMetaSizeV1

	PowMasterMsgStatusContinue fields.VarInt1 = 0
	PowMasterMsgStatusSuccess  fields.VarInt1 = 1
	PowMasterMsgStatusStop     fields.VarInt1 = 2
	PowMasterMsgStatusError    fields.VarInt1 = 3
)

type PowMasterMsg struct {
	Status         fields.VarInt1 //
	CoinbaseMsgNum fields.VarInt4
	NonceBytes     fields.Bytes4
	BlockHeadMeta  interfaces.Block
}

func NewPowMasterMsg() *PowMasterMsg {
	return &PowMasterMsg{
		0, 0, []byte{0, 0, 0, 0}, blocks.NewEmptyBlock_v1(nil),
	}
}

func (p *PowMasterMsg) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	b1, _ := p.Status.Serialize()
	b2, _ := p.CoinbaseMsgNum.Serialize()
	b3, _ := p.NonceBytes.Serialize()
	b4, _ := p.BlockHeadMeta.SerializeExcludeTransactions()
	buf.Write(b1)
	buf.Write(b2)
	buf.Write(b3)
	buf.Write(b4)
	return buf.Bytes(), nil
}

func (p *PowMasterMsg) Parse(buf []byte, seek uint32) (uint32, error) {
	if uint32(len(buf))-seek < PowMasterMsgSize {
		return 0, fmt.Errorf("size error.")
	}
	seek, _ = p.Status.Parse(buf, seek)
	seek, _ = p.CoinbaseMsgNum.Parse(buf, seek)
	seek, _ = p.NonceBytes.Parse(buf, seek)
	seek, _ = p.BlockHeadMeta.ParseExcludeTransactions(buf, seek)
	return seek, nil
}

func (p *PowMasterMsg) Size() uint32 {
	return PowMasterMsgSize
}
