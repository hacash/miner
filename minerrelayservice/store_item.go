package minerrelayservice

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/mint/difficulty"
)

type StoreItemUserMiningResult struct {
	BodyVersion               uint8 // 版本号
	IsMintSuccessed           uint8 // 是否挖掘成功
	IsSaveMiningResultHash    uint8 // 是否保存hash
	MiningResultHash          fields.Hash
	IsSaveMiningResultNonce   uint8 // 是否保存 nonce
	MiningResultHeadNonce     fields.Bytes4
	MiningResultCoinbaseNonce fields.Hash

	// cache data
	blockHeight   uint64
	rewardAddress fields.Address
}

func NewStoreItemUserMiningResultV0() *StoreItemUserMiningResult {
	return &StoreItemUserMiningResult{
		BodyVersion:             0,
		IsMintSuccessed:         0,
		IsSaveMiningResultHash:  0,
		IsSaveMiningResultNonce: 0,
		blockHeight:             0,
		rewardAddress:           nil,
	}
}

// json api
func (s *StoreItemUserMiningResult) Describe() map[string]interface{} {
	var data = map[string]interface{}{}
	data["mint_success"] = s.IsMintSuccessed
	if s.IsSaveMiningResultHash == 1 {
		data["result_hash"] = s.MiningResultHash.ToHex()
		userblkhei := uint64(1)
		if s.blockHeight > 1 {
			userblkhei = s.blockHeight
		}
		data["result_hash_worth"] = difficulty.CalculateHashWorth(userblkhei, s.MiningResultHash).String()
	}
	if s.IsSaveMiningResultNonce == 1 {
		hdnc, _ := s.MiningResultHeadNonce.Serialize()
		data["head_nonce"] = hex.EncodeToString(hdnc)
		data["coinbase_nonce"] = s.MiningResultCoinbaseNonce.ToHex()
	}
	if s.blockHeight > 1 {
		data["block_height"] = s.blockHeight
	}
	if s.rewardAddress != nil {
		data["reward_address"] = s.rewardAddress.ToReadable()
	}
	return data
}

func (s *StoreItemUserMiningResult) Parse(buf []byte, seek uint32) (uint32, error) {
	if uint32(len(buf)) < seek+4 {
		return 0, fmt.Errorf("buf len too small")
	}
	s.BodyVersion = buf[seek]
	seek += 1
	s.IsMintSuccessed = buf[seek]
	seek += 1
	s.IsSaveMiningResultHash = buf[seek]
	seek += 1
	if s.IsSaveMiningResultHash == 1 {
		s.MiningResultHash = make([]byte, 32)
		copy(s.MiningResultHash, buf[seek:seek+32])
		seek += 32
	}
	s.IsSaveMiningResultNonce = buf[seek]
	seek += 1
	if s.IsSaveMiningResultNonce == 1 {
		s.MiningResultHeadNonce = make([]byte, 4)
		copy(s.MiningResultHeadNonce, buf[seek:seek+4])
		seek += 4
		s.MiningResultCoinbaseNonce = make([]byte, 32)
		copy(s.MiningResultCoinbaseNonce, buf[seek:seek+32])
		seek += 32
	}
	return seek, nil
}

func (s *StoreItemUserMiningResult) Serialize() []byte {
	buf := bytes.NewBuffer([]byte{s.BodyVersion})
	buf.Write([]byte{s.IsMintSuccessed})
	buf.Write([]byte{s.IsSaveMiningResultHash})
	if s.IsSaveMiningResultHash == 1 {
		buf.Write(s.MiningResultHash)
	}
	buf.Write([]byte{s.IsSaveMiningResultNonce})
	if s.IsSaveMiningResultNonce == 1 {
		buf.Write(s.MiningResultHeadNonce)
		buf.Write(s.MiningResultCoinbaseNonce)
	}
	return buf.Bytes()
}
