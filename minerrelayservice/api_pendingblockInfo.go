package minerrelayservice

import (
	"net/http"
	"time"
)

// 当前正在挖掘的区块信息
func (api *RelayService) pendingBlockInfo(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	if api.penddingBlockStuff == nil {
		ResponseErrorString(w, "pending block not yet")
		return
	}

	// 等待目标区块
	waitTargetHeight := CheckParamUint64(r, "wait_10sec_for_block_height", 0)
	if waitTargetHeight > 0 {
		waitSec := 0         // 已经等待秒
		waitMaxTimeout := 10 // 等待 10 秒// CheckParamUint64(r, "wait_timeout", 3)
		// 开始等待
		for {
			if api.penddingBlockStuff != nil {
				pendingHeight := api.penddingBlockStuff.BlockHeadMeta.GetHeight()
				if pendingHeight == waitTargetHeight {
					break // 等待成功！！！
				}
				if waitTargetHeight < pendingHeight {
					// 高度小于目前，不可能等待到了
					ResponseErrorString(w, "wait target block height less than pending height")
					return
				}
				// 继续等待
			}
			if waitSec >= waitMaxTimeout {
				// 等待结束
				ResponseErrorString(w, "wait target block timeout")
				return
			}
			// 等待一秒钟
			waitSec++
			time.Sleep(time.Second)
		}
	}

	// 开始返回全部信息

	// mei
	isUnitMei := CheckParamBool(r, "unitmei", false)
	isOnlyReturnHeight := CheckParamBool(r, "only_height", false)

	cblk := api.penddingBlockStuff.BlockHeadMeta
	cbtx := api.penddingBlockStuff.CoinbaseTx

	// 仅仅返回区块高度
	// 用于 wait_block_height 判断下一个挖矿区块已经到来
	if isOnlyReturnHeight {
		blockinfo := make(map[string]interface{})
		blockinfo["height"] = cblk.GetHeight()
		data := ResponseCreateData("block", blockinfo)
		ResponseData(w, data)
		return // 返回高度
	}

	// 返回所有详细信息

	// return
	blockinfo := make(map[string]interface{})
	blockinfo["version"] = cblk.Version() // 版本号
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
	mrklhashs := api.penddingBlockStuff.MrklRelatedTreeListForCoinbaseTxModify
	mrklshows := make([]string, len(mrklhashs))
	for i, v := range mrklhashs {
		mrklshows[i] = v.ToHex()
	}

	// set data
	data["mrkl_miner_related_hash_list"] = mrklshows

	// ok
	ResponseData(w, data)

}
