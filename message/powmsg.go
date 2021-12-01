package message

import (
	"bytes"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
)

const (
	PowMasterMsgSize = blocks.BlockHeadSize + blocks.BlockMetaSizeV1 + 1 + 4 + 4

	PowMasterMsgStatusContinue                          fields.VarUint1 = 0
	PowMasterMsgStatusSuccess                           fields.VarUint1 = 1
	PowMasterMsgStatusStop                              fields.VarUint1 = 2
	PowMasterMsgStatusMostPowerHash                     fields.VarUint1 = 3
	PowMasterMsgStatusMostPowerHashAndRequestNextMining fields.VarUint1 = 4
)

type PowMasterMsg struct {
	Status         fields.VarUint1 //
	CoinbaseMsgNum fields.VarUint4
	NonceBytes     fields.Bytes4
	BlockHeadMeta  interfaces.Block
}

func NewPowMasterMsg() *PowMasterMsg {
	return &PowMasterMsg{
		0, 0, []byte{0, 0, 0, 0}, blocks.NewEmptyBlockVersion1(nil),
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
	var e error = nil
	if uint32(len(buf))-seek < PowMasterMsgSize {
		return 0, fmt.Errorf("size error.")
	}
	seek, e = p.Status.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = p.CoinbaseMsgNum.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = p.NonceBytes.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	p.BlockHeadMeta, seek, e = blocks.ParseExcludeTransactions(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (p *PowMasterMsg) Size() uint32 {
	return PowMasterMsgSize
}
