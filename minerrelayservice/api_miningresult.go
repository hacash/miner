package minerrelayservice

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/miner/message"
	"github.com/hacash/mint/difficulty"
	"net/http"
)

// 查询挖矿结果
func (api *RelayService) queryMiningResult(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	if api.ldb == nil {
		ResponseErrorString(w, "store database not init.")
		return
	}

	onlyWorth := CheckParamBool(r, "only_worth", false)

	idxStart := CheckParamUint64(r, "idx_start", 0)
	idxLimit := CheckParamUint64(r, "idx_limit", 0)
	if idxLimit > 50 {
		// 通过索引查找，最多50条
		idxLimit = 50
	}

	// 查询的 key
	var queryKeys = make([][]byte, 0)
	var queryHeis = make([]uint64, 0)
	var queryAddrs = make([]fields.Address, 0)

	if idxLimit > 0 {
		// 通过自增索引查询
		max := idxStart + idxLimit
		for i := idxStart; i < max; i++ {
			kob1 := fields.VarUint5(i)
			idxkey, _ := kob1.Serialize()
			k1 := []byte("mri" + string(idxkey))
			if v1, e1 := api.ldb.Get(k1, nil); e1 == nil {
				kk := []byte("mr" + string(v1))
				queryKeys = append(queryKeys, kk) // append
				hei := fields.VarUint5(0)
				hei.Parse(v1, 0)
				queryHeis = append(queryHeis, uint64(hei)) // append
				rwdaddr := fields.Address{}
				rwdaddr.Parse(v1, 5)
				queryAddrs = append(queryAddrs, rwdaddr) // append
			}
		}
	} else {
		// 通过区块高度和奖励地址组成的可以查询
		addrstr := CheckParamString(r, "reward_address", "")
		rwdaddr, e1 := fields.CheckReadableAddress(addrstr)
		if e1 != nil {
			ResponseErrorString(w, "reward address format error")
			return
		}
		blkheiStart := CheckParamUint64(r, "block_height_start", 0)
		blkheiLimit := CheckParamUint64(r, "block_height_limit", 0)
		if blkheiLimit > 50 {
			// 通过索引查找，最多50条
			blkheiLimit = 50
		}
		if blkheiLimit == 0 {
			data := ResponseCreateData("list", []int{})
			ResponseData(w, data)
			return
		}
		max := blkheiStart + blkheiLimit
		for i := blkheiStart; i < max; i++ {
			kob2 := fields.VarUint5(i)
			heikey, _ := kob2.Serialize()
			keybuf := bytes.NewBuffer(heikey)
			keybuf.Write(*rwdaddr)
			keybts := keybuf.Bytes()
			k1 := []byte("mr" + string(keybts))
			queryKeys = append(queryKeys, k1)         // append
			queryHeis = append(queryHeis, i)          // append
			queryAddrs = append(queryAddrs, *rwdaddr) // append
		}
	}

	// 返回数组
	retlist := make([]interface{}, 0)

	// 返回空
	if len(queryKeys) == 0 {
		//fmt.Println("len(queryKeys) == 0")
		data := ResponseCreateData("list", retlist)
		ResponseData(w, data)
		return
	}

	// 查询数据
	queryObjs := make([]*StoreItemUserMiningResult, 0)
	for i, key := range queryKeys {
		//fmt.Println(key)
		if v1, e1 := api.ldb.Get(key, nil); e1 == nil {
			var stoitem = &StoreItemUserMiningResult{}
			if _, e2 := stoitem.Parse(v1, 0); e2 == nil {
				stoitem.blockHeight = queryHeis[i]
				stoitem.rewardAddress = queryAddrs[i]
				queryObjs = append(queryObjs, stoitem)
			}
		}
	}

	// 返回空
	if len(queryObjs) == 0 {
		//fmt.Println("len(queryObjs) == 0")
		data := ResponseCreateData("list", retlist)
		ResponseData(w, data)
		return
	}

	// 解析数组，并计算 hash worth
	for _, obj := range queryObjs {
		des := obj.Describe()
		if onlyWorth {
			delete(des, "head_nonce")
			delete(des, "coinbase_nonce")
			delete(des, "mint_success")
			delete(des, "result_hash")
			if idxLimit == 0 {
				delete(des, "reward_address")
			}
		}
		retlist = append(retlist, des)
	}

	// ok
	data := ResponseCreateData("list", retlist)
	ResponseData(w, data)
}

// 提交挖矿结果
func (api *RelayService) submitMiningResult(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

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

	// 返回值
	var retdata = map[string]interface{}{}

	// 检查挖矿是否成功
	var isRealMintSuccess = false
	newstuff, newhx := tarstuff.CalculateBlockHashByBothNonce(headNonce, coinbaseNonce, true)
	if isMintSuccess {
		// 上报挖掘成功消息
		// 判断哈希满足要求
		newblock := newstuff.GetHeadMetaBlock()
		if difficulty.CheckHashDifficultySatisfyByBlock(newhx, newblock) {
			rptmsg := message.MsgReportMiningResult{
				MintSuccessed: 1,
				BlockHeight:   fields.VarUint5(blkhei),
				HeadNonce:     headNonce,
				CoinbaseNonce: coinbaseNonce,
			}
			// 发送消息
			message.MsgSendToTcpConn(api.service_tcp, message.MinerWorkMsgTypeReportMiningResult, rptmsg.Serialize())
			isRealMintSuccess = true
		}
		// 不满足难度要求，什么都不干，
	}

	// 统计算力
	go api.saveMiningResultToStore(*rwdaddr, isRealMintSuccess, newstuff)

	// ok
	retdata["stat"] = "ok"
	ResponseData(w, retdata)

}
