package minerrelayservice

import (
	"bytes"
	"github.com/hacash/core/fields"
	"net/http"
)

// Query mining results
func (api *RelayService) queryMiningResult(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	if api.ldb == nil {
		ResponseErrorString(w, "store database not init.")
		return
	}

	onlyWorth := CheckParamBool(r, "only_worth", false)

	idxStart := CheckParamUint64(r, "idx_start", 0)
	idxLimit := CheckParamUint64(r, "idx_limit", 0)
	if idxLimit > 50 {
		// Search by index, up to 50
		idxLimit = 50
	}

	// Query key
	var queryKeys = make([][]byte, 0)
	var queryHeis = make([]uint64, 0)
	var queryAddrs = make([]fields.Address, 0)

	if idxLimit > 0 {
		// Query by auto incrementing index
		max := idxStart + idxLimit
		for i := idxStart; i < max; i++ {
			kob1 := fields.VarUint5(i)
			idxkey, _ := kob1.Serialize()
			k1 := []byte("mri" + string(idxkey))
			if v1, e1 := api.ldb.Get(k1, nil); e1 == nil {
				kk := []byte("mr" + string(v1))
				queryKeys = append(queryKeys, kk) // append
				hei := fields.BlockHeight(0)
				hei.Parse(v1, 0)
				queryHeis = append(queryHeis, uint64(hei)) // append
				rwdaddr := fields.Address{}
				rwdaddr.Parse(v1, 5)
				queryAddrs = append(queryAddrs, rwdaddr) // append
			}
		}
	} else {
		// It can be queried through the block height and reward address
		addrstr := CheckParamString(r, "reward_address", "")
		rwdaddr, e1 := fields.CheckReadableAddress(addrstr)
		if e1 != nil {
			ResponseErrorString(w, "reward address format error")
			return
		}
		blkheiStart := CheckParamUint64(r, "block_height_start", 0)
		blkheiLimit := CheckParamUint64(r, "block_height_limit", 0)
		if blkheiLimit > 50 {
			// Search by index, up to 50
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

	// Return array
	retlist := make([]interface{}, 0)

	// Return null
	if len(queryKeys) == 0 {
		//fmt.Println("len(queryKeys) == 0")
		data := ResponseCreateData("list", retlist)
		ResponseData(w, data)
		return
	}

	// Query data
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

	// Return null
	if len(queryObjs) == 0 {
		//fmt.Println("len(queryObjs) == 0")
		data := ResponseCreateData("list", retlist)
		ResponseData(w, data)
		return
	}

	// Parse the array and calculate hash worth
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

// Submit mining results
func (api *RelayService) submitMiningResult(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	// Mining success
	//isMintSuccess := CheckParamBool(r, "mint_success", false)

	// Record reward address
	var e1 error
	//addrstr := CheckParamString(r, "reward_address", "")
	//rwdaddr, e1 := fields.CheckReadableAddress(addrstr)
	if e1 != nil {
		ResponseErrorString(w, "reward address format error")
		return
	}

	blkhei := CheckParamUint64(r, "block_height", 0)
	if blkhei == 0 {
		ResponseErrorString(w, "block_height must give")
		return
	}

	// Return value
	var retdata = map[string]interface{}{}

	/*
		// seek
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

		// Check whether the mining is successful
		var isRealMintSuccess = false
		newstuff, newhx := tarstuff.CalculateBlockHashByBothNonce(headNonce, coinbaseNonce, true)
		if isMintSuccess {
			// Report mining success message
			// Judge whether the hash meets the requirements
			newblock := newstuff.GetHeadMetaBlock()
			if difficulty.CheckHashDifficultySatisfyByBlock(newhx, newblock) {
				rptmsg := message.MsgReportMiningResult{
					MintSuccessed: 1,
					BlockHeight:   fields.BlockHeight(blkhei),
					HeadNonce:     headNonce,
					CoinbaseNonce: coinbaseNonce,
				}
				// send message
				message.MsgSendToTcpConn(api.service_tcp, message.MinerWorkMsgTypeReportMiningResult, rptmsg.Serialize())
				isRealMintSuccess = true
			}
			// 不满足难度要求，什么都不干，
		}

		// Statistical computing power
		go api.saveMiningResultToStore(*rwdaddr, isRealMintSuccess, newstuff)


	*/
	// ok
	retdata["stat"] = "ok"
	ResponseData(w, retdata)

}
