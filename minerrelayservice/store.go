package minerrelayservice

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/hacash/chain/leveldb"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/sys"
	"github.com/hacash/miner/message"
	"os"
)

const KeyAutoIdxUserMiningResult = "autoidx1"

// 初始化存储
func (api *RelayService) initStore() {

	if api.config.StoreEnable == false {
		return
	}

	// 初始化
	absDataDir := sys.AbsDir(api.config.DataDir)
	err := os.MkdirAll(absDataDir, os.ModePerm)
	if err != nil {
		fmt.Println("[Miner Relay Service] initStore ERROR:")
		fmt.Println(err)
		return
	}

	// create db
	ldb, e2 := leveldb.OpenFile(absDataDir, nil)
	if e2 != nil {
		fmt.Println("[Miner Relay Service] leveldb.OpenFile ERROR:")
		fmt.Println(e2)
		return
	}

	api.ldb = ldb

	// 读取自增索引
	idx1bts, e3 := ldb.Get([]byte(KeyAutoIdxUserMiningResult), nil)
	if e3 != nil {
		idx1bts = []byte{0, 0, 0, 0, 0, 0, 0, 0}
	}
	idx1 := binary.BigEndian.Uint64(idx1bts)

	// 自增索引
	api.userMiningResultStoreAutoIdx = idx1

}

// 储存用户挖矿统计
func (api *RelayService) saveMiningResultToStore(rwdaddr fields.Address, isMintSuccessed bool, resultStuff *message.MsgPendingMiningBlockStuff) {

	//fmt.Println("saveMiningResultToStore start")

	// 串行保存
	api.userMiningResultStoreAutoIdxMutex.Lock()
	defer api.userMiningResultStoreAutoIdxMutex.Unlock()

	if api.ldb == nil || api.config.StoreEnable == false {
		return
	}

	//fmt.Println("NewStoreItemUserMiningResultV0")
	// 组件储存单元
	stoitem := NewStoreItemUserMiningResultV0()
	if isMintSuccessed {
		stoitem.IsMintSuccessed = 1
	}
	if api.config.SaveMiningHash {
		stoitem.IsSaveMiningResultHash = 1
		stoitem.MiningResultHash = resultStuff.BlockHeadMeta.HashFresh()
	}
	if api.config.SaveMiningNonce {
		stoitem.IsSaveMiningResultNonce = 1
		stoitem.MiningResultHeadNonce = resultStuff.BlockHeadMeta.GetNonceByte()
		stoitem.MiningResultCoinbaseNonce = resultStuff.CoinbaseTx.MinerNonce
	}
	stobts := stoitem.Serialize()

	// key
	kob1 := fields.VarUint5(resultStuff.BlockHeadMeta.GetHeight())
	heikey, _ := kob1.Serialize()
	keybuf := bytes.NewBuffer(heikey)
	keybuf.Write(rwdaddr)

	// save item
	keybts := keybuf.Bytes()
	k1 := []byte("mr" + string(keybts))
	// check k1
	_, e1 := api.ldb.Get(k1, nil)
	if e1 == nil {
		//fmt.Println("if e1 == nil { api.ldb.Put(k1, stobts, nil)")
		// 已经存在，则只需要替换
		api.ldb.Put(k1, stobts, nil) // mr
		return                       // 替换了即返回
	}

	//fmt.Println("api.ldb.Put(k1, stobts, nil)")
	// 保存新的
	api.ldb.Put(k1, stobts, nil) // mr

	// save idx
	api.userMiningResultStoreAutoIdx += 1
	kob2 := fields.VarUint5(api.userMiningResultStoreAutoIdx)
	idxkey, _ := kob2.Serialize()
	//fmt.Println("mri"+string(idxkey), `api.ldb.Put([]byte("mri"+string(idxkey)), keybts, nil) // mri`)
	api.ldb.Put([]byte("mri"+string(idxkey)), keybts, nil) // mri

	// idx
	kob3 := fields.VarUint8(api.userMiningResultStoreAutoIdx)
	idxkeyauto, _ := kob3.Serialize()
	//fmt.Println(KeyAutoIdxUserMiningResult, `api.ldb.Put([]byte(KeyAutoIdxUserMiningResult), idxkeyauto, nil)`)
	api.ldb.Put([]byte(KeyAutoIdxUserMiningResult), idxkeyauto, nil)

	// save ok
}

// 储存挖矿stuff
func (api *RelayService) saveMiningBlockStuffToStore(stuff *message.MsgPendingMiningBlockStuff) {

	if api.ldb == nil || api.config.StoreEnable == false || api.config.SaveMiningBlockStuff == false {
		return
	}

	// 储存
	blkhei := stuff.BlockHeadMeta.GetHeight()
	stodatas := stuff.Serialize()
	k1 := fields.VarUint5(blkhei)
	heikey, _ := k1.Serialize()

	// save
	err := api.ldb.Put(heikey, stodatas, nil)
	if err != nil {
		fmt.Println("[Miner Relay Service] saveMiningBlockStuffToStore ERROR:")
		fmt.Println(err)
	}
}

// 读取储存挖矿stuff
func (api *RelayService) readMiningBlockStuffFormStore(blkhei uint64) *message.MsgPendingMiningBlockStuff {

	if api.ldb == nil || api.config.StoreEnable == false || api.config.SaveMiningBlockStuff == false {
		return nil
	}

	// 储存
	k1 := fields.VarUint5(blkhei)
	heikey, _ := k1.Serialize()

	// save
	stodatas, err := api.ldb.Get(heikey, nil)
	if err != nil {
		fmt.Println("[Miner Relay Service] readMiningBlockStuffFormStore ERROR:")
		fmt.Println(err)
	}

	var stuff = message.MsgPendingMiningBlockStuff{}
	_, e2 := stuff.Parse(stodatas, 0)
	if e2 != nil {
		fmt.Println("[Miner Relay Service] MsgPendingMiningBlockStuff.Parse ERROR:")
		fmt.Println(e2)
	}

	// ok
	return &stuff
}
