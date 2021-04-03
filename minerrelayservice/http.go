package minerrelayservice

import (
	"io/ioutil"
	"net/http"
)

func (api *RelayService) initRoutes() {

	// query
	api.queryRoutes["pending_block_info"] = api.pendingBlockInfo // 查询当前正在挖掘的区块信息
	api.queryRoutes["mining_stuff"] = api.miningStuff            // 请求挖矿数据

	// submit
	api.submitRoutes["mining_result"] = api.miningResult // 提交挖矿结果

	//// create
	//api.createRoutes["accounts"] = api.createAccounts
	//api.createRoutes["value_transfer_tx"] = api.createValueTransferTx
	//
	//// query
	//api.queryRoutes["total_supply"] = api.totalSupply
	//
	//api.queryRoutes["balances"] = api.balances
	//api.queryRoutes["diamond"] = api.diamond
	//
	//api.queryRoutes["last_block"] = api.lastBlock
	//api.queryRoutes["block_intro"] = api.blockIntro
	//
	//api.queryRoutes["scan_value_transfers"] = api.scanTransfersOfTransactionByPosition
	//
	//// operate
	//api.operateRoutes["raise_tx_fee"] = api.raiseTxFee

}

func (api *RelayService) dealQuery(w http.ResponseWriter, r *http.Request) {
	api.dealRoutes(api.queryRoutes, w, r, false)
}

func (api *RelayService) dealCreate(w http.ResponseWriter, r *http.Request) {
	api.dealRoutes(api.createRoutes, w, r, false)
}

func (api *RelayService) dealSubmit(w http.ResponseWriter, r *http.Request) {
	api.dealRoutes(api.submitRoutes, w, r, true)
}

func (api *RelayService) dealOperate(w http.ResponseWriter, r *http.Request) {
	api.dealRoutes(api.operateRoutes, w, r, false)
}

func (api *RelayService) dealCalculate(w http.ResponseWriter, r *http.Request) {
	api.dealRoutes(api.calculateRoutes, w, r, false)
}

func (api *RelayService) dealRoutes(routes map[string]func(*http.Request, http.ResponseWriter, []byte), w http.ResponseWriter, r *http.Request, gotbodybytes bool) {
	var err error
	var bodybytes []byte = nil

	if gotbodybytes {
		bodybytes, err = ioutil.ReadAll(r.Body)
		if err != nil {
			ResponseError(w, err)
			return
		}
	}

	r.ParseForm()

	actionName := r.FormValue("action")

	if len(actionName) == 0 {
		ResponseErrorString(w, "param 'action' must give.")
		return
	}

	action, actok := routes[actionName]
	if !actok {
		ResponseErrorString(w, "not find action <"+actionName+">.")
		return
	}

	// call action
	action(r, w, bodybytes)
}
