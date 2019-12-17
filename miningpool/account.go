package miningpool

import (
	"github.com/hacash/chain/mapset"
	"github.com/hacash/core/fields"
	"math/big"
)

type Account struct {
	address fields.Address // 获得奖励地址

	activeClients mapset.Set // [*Client] // 正在连接的客户端

	periodStartBlockHeight uint64 // 统计周期开始区块

	realtimePowWorth *big.Int // 周期内算力统计

	///////////////////////////////////////////////

	FindBlocks              fields.VarInt4 // 挖出的区块数量
	FindCoins               fields.VarInt4 // 挖出的币数量
	CompleteRewards         fields.Amount  // 已完成并打币的奖励  单位： ㄜ240  （10^8）
	DeservedRewards         fields.Amount  // 应得但还没有打币的奖励  单位： ㄜ240  （10^8）
	PrevTransferBlockHeight fields.VarInt4 // 上一次打币时的区块

}
