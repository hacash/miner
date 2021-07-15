package minerrelayservice

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"github.com/hacash/core/blocks"
	"net/http"
)

// 返回可供挖矿使用的数据
func (api *RelayService) miningStuff(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	stfobj := api.penddingBlockStuff

	if stfobj == nil {
		ResponseErrorString(w, "pending block not yet")
		return
	}

	// 两个nonce
	headnonce := bytes.Repeat([]byte{0}, 4)
	coinbasenonce := make([]byte, 32)
	cmns := CheckParamString(r, "coinbase_nonce", "")
	cbts, e1 := hex.DecodeString(cmns)
	if e1 != nil || len(cbts) != 32 {
		rand.Read(coinbasenonce) // 随机生成nonce
	} else {
		coinbasenonce = cbts // 使用传递的nonce
	}

	// 计算填充
	newstuff, _ := stfobj.CalculateBlockHashByBothNonce(headnonce, coinbasenonce, true)

	// 计算 挖矿 stuff
	stuffbytes := blocks.CalculateBlockHashBaseStuff(newstuff.BlockHeadMeta)

	// 返回
	data := ResponseCreateData("stuff", hex.EncodeToString(stuffbytes))
	data["coinbase_nonce"] = hex.EncodeToString(coinbasenonce)
	data["head_nonce_start"] = 79
	data["head_nonce_len"] = 4
	data["height"] = newstuff.BlockHeadMeta.GetHeight()

	// ok
	ResponseData(w, data)

}
