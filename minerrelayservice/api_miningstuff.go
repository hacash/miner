package minerrelayservice

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/hacash/core/blocks"
	"github.com/hacash/mint/difficulty"
	"net/http"
)

// Return data available for mining
func (api *RelayService) miningStuff(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	stfobj := api.penddingBlockStuff

	if stfobj == nil {
		ResponseErrorString(w, "pending block not yet")
		return
	}

	// Two nonces
	//headnonce := bytes.Repeat([]byte{0}, 4)
	coinbasenonce := make([]byte, 32)
	cmns := CheckParamString(r, "coinbase_nonce", "")
	cbts, e1 := hex.DecodeString(cmns)
	if e1 != nil || len(cbts) != 32 {
		rand.Read(coinbasenonce) // Random generation nonce
	} else {
		coinbasenonce = cbts // Use passed nonce
	}

	// replace
	cbtx := stfobj.CoinbaseTx.CopyForMining()
	cbtx.MinerNonce = coinbasenonce
	var cb_hash = cbtx.Hash()
	// mklrRoot
	var useblk = stfobj.BlockHeadMeta.CopyHeadMetaForMining()
	mklr_root := blocks.CalculateMrklRootByCoinbaseTxModify(cb_hash, stfobj.MrklCheckTreeList)
	useblk.SetMrklRoot(mklr_root)

	//// Calculate fill
	//newblk, _ := stfobj.CalculateBlockHashByBothNonce(headnonce, coinbasenonce, true)

	// Calculation of mining stuff
	stuffbytes := blocks.CalculateBlockHashBaseStuff(useblk)

	// return
	height := useblk.GetHeight()
	data := ResponseCreateData("stuff", hex.EncodeToString(stuffbytes))
	data["coinbase_nonce"] = hex.EncodeToString(coinbasenonce)
	data["head_nonce_start"] = 79
	data["head_nonce_len"] = 4
	data["height"] = height
	data["target_difficulty_hash"] = hex.EncodeToString(difficulty.Uint32ToHash(height, useblk.GetDifficulty()))

	// ok
	ResponseData(w, data)

}
