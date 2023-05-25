package minerrelayservice

import (
	"fmt"
	interfaces2 "github.com/hacash/miner/interfaces"
	"net/http"
	"time"
)

// Historical block information
func (api *RelayService) readHistoricalMiningBlockInfo(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	if api.ldb == nil {
		ResponseErrorString(w, "config [store] enable not open")
		return
	}

	// mei
	isUnitMei := CheckParamBool(r, "unitmei", false)
	blkHeight := CheckParamUint64(r, "height", 0)
	if blkHeight == 0 {
		ResponseErrorString(w, "height must give")
		return
	}

	// read
	stuff := api.readMiningBlockStuffFormStore(blkHeight)
	if stuff == nil {
		ResponseError(w, fmt.Errorf("not find height %d", blkHeight))
		return
	}

	// return
	returnStuff(w, stuff, isUnitMei)
}

// Currently mining block information
func (api *RelayService) pendingBlockInfo(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	if api.penddingBlockStuff == nil {
		ResponseErrorString(w, "pending block not yet")
		return
	}

	// Waiting for target block
	waitTargetHeight := CheckParamUint64(r, "wait_10sec_for_block_height", 0)
	if waitTargetHeight > 0 {
		waitSec := 0         // 已经等待秒
		waitMaxTimeout := 10 // 等待 10 秒// CheckParamUint64(r, "wait_timeout", 3)
		// Start waiting
		for {
			if api.penddingBlockStuff != nil {
				pendingHeight := api.penddingBlockStuff.BlockHeadMeta.GetHeight()
				if pendingHeight == waitTargetHeight {
					break // Waiting for success!!!
				}
				if waitTargetHeight < pendingHeight {
					// The height is lower than the current height, so it is impossible to wait
					ResponseErrorString(w, "wait target block height less than pending height")
					return
				}
				// Keep waiting
			}
			if waitSec >= waitMaxTimeout {
				// Wait for end
				ResponseErrorString(w, "wait target block timeout")
				return
			}
			// Wait one second
			waitSec++
			time.Sleep(time.Second)
		}
	}

	// Start returning all information

	// mei
	isUnitMei := CheckParamBool(r, "unitmei", false)
	isOnlyReturnHeight := CheckParamBool(r, "only_height", false)

	cblk := api.penddingBlockStuff.BlockHeadMeta

	// Return block height only
	// For wait_ block_ Height judges that the next mining block has arrived
	if isOnlyReturnHeight {
		blockinfo := make(map[string]interface{})
		blockinfo["height"] = cblk.GetHeight()
		data := ResponseCreateData("block", blockinfo)
		ResponseData(w, data)
		return // Return height
	}

	// return
	returnStuff(w, api.penddingBlockStuff, isUnitMei)

}

// Return all details
func returnStuff(w http.ResponseWriter, stuff *interfaces2.PoWStuffOverallData, isUnitMei bool) {

	cblk := stuff.BlockHeadMeta
	cbtx := stuff.CoinbaseTx

	// return
	blockinfo := make(map[string]interface{})
	blockinfo["version"] = cblk.Version() // Version number
	blockinfo["height"] = cblk.GetHeight()
	blockinfo["timestamp"] = cblk.GetTimestamp()
	blockinfo["prevhash"] = cblk.GetPrevHash().ToHex()
	blockinfo["transaction_count"] = cblk.GetTransactionCount()
	blockinfo["difficulty"] = cblk.GetDifficulty()
	blockinfo["witness_stage"] = cblk.GetWitnessStage()

	// data
	data := ResponseCreateData("block", blockinfo)

	// coinbase
	// set data
	data["coinbase"] = cbtx.Describe(isUnitMei, true)

	// Mrkl hash
	mrklhashs := stuff.MrklCheckTreeList
	mrklshows := make([]string, len(mrklhashs))
	for i, v := range mrklhashs {
		mrklshows[i] = v.ToHex()
	}

	// set data
	data["mrkl_miner_related_hash_list"] = mrklshows

	// ok
	ResponseData(w, data)

}
