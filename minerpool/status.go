package minerpool

import (
	"bytes"
	"fmt"
	"github.com/hacash/core/fields"
)

const (
	MinerPoolStatusSize = 32 * 20
)

type MinerPoolStatus struct {
	FindBlocks                            fields.VarInt4 // 挖出的区块数量
	FindCoins                             fields.VarInt4 // 挖出的币数量
	FindBlockHashHeightTableLastestNumber fields.VarInt4 // 挖出的区块id表最新值

}

func (s *MinerPoolStatus) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	b1, _ := s.FindBlocks.Serialize()
	b2, _ := s.FindCoins.Serialize()
	b3, _ := s.FindBlockHashHeightTableLastestNumber.Serialize()
	buf.Write(b1)
	buf.Write(b2)
	buf.Write(b3)
	resbuf := make([]byte, MinerPoolStatusSize)
	copy(resbuf, buf.Bytes())
	return resbuf, nil
}

func (s *MinerPoolStatus) Parse(buf []byte, seek uint32) (uint32, error) {
	if uint32(len(buf))-seek < 4+4+4 {
		return 0, fmt.Errorf("size error.")
	}
	seek, _ = s.FindBlocks.Parse(buf, seek)
	seek, _ = s.FindCoins.Parse(buf, seek)
	seek, _ = s.FindBlockHashHeightTableLastestNumber.Parse(buf, seek)
	return seek, nil
}

func (s *MinerPoolStatus) Size() uint32 {
	return MinerPoolStatusSize
}
