package minerrelayservice

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/fields"
	"net/http"
)

// 提交挖矿结果
func (api *RelayService) miningResult(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	// 是否挖矿成功
	isMintSuccess := CheckParamBool(r, "mint_success", false)

	// 记录奖励地址
	addrstr := CheckParamString(r, "reward_address", "")
	rwdaddr, e1 := fields.CheckReadableAddress(addrstr)
	if e1 != nil {
		ResponseErrorString(w, "reward address format error")
		return
	}

	blkhei := CheckParamUint64(r, "block_height", 0)
	if blkhei == 0 {
		ResponseErrorString(w, "block_height must give")
		return
	}

	// 寻找
	tarstuff := api.checkoutMiningStuff(blkhei)
	if tarstuff == nil {
		ResponseError(w, fmt.Errorf("not find mining stuff of block height %d", blkhei))
		return
	}

	// head nonce
	bts1 := CheckParamString(r, "head_nonce", "")
	headNonce, ehn := hex.DecodeString(bts1)
	if ehn != nil || len(headNonce) != 4 {
		ResponseErrorString(w, "head_nonce format error")
		return
	}

	// coinbase nonce
	bts2 := CheckParamString(r, "coinbase_nonce", "")
	coinbaseNonce, ecbn := hex.DecodeString(bts2)
	if ecbn != nil || len(coinbaseNonce) != 32 {
		ResponseErrorString(w, "coinbase_nonce format error")
		return
	}

	// TODO: 检查挖矿是否成功

	data := ResponseCreateData("stat", isMintSuccess)
	data["rwdaddr"] = rwdaddr.ToReadable()

	// ok
	ResponseData(w, data)

}
