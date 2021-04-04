package minerrelayservice

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/fields"
	"github.com/hacash/miner/message"
	"github.com/hacash/mint/difficulty"
	"math/rand"
	"net"
)

type ConnClient struct {
	server  *RelayService
	id      uint64
	conn    *net.TCPConn
	rwdaddr fields.Address // 奖励地址
}

func NewConnClient(server *RelayService, conn *net.TCPConn, rwdaddr fields.Address) *ConnClient {
	cid := rand.Uint64()
	return &ConnClient{
		server:  server,
		id:      cid,
		conn:    conn,
		rwdaddr: rwdaddr,
	}
}

// Handle
func (m *ConnClient) Handle() error {
	for {
		// 读取消息
		msgty, msgbody, err := message.MsgReadFromTcpConn(m.conn, 0)
		if err != nil {
			return err
		}
		// 解析消息
		if msgty == message.MinerWorkMsgTypeReportMiningResult {
			var result = message.MsgReportMiningResult{}
			_, err := result.Parse(msgbody, 0)
			if err != nil {
				return err
			}
			// 处理
			pickupstuff := m.server.checkoutMiningStuff(uint64(result.BlockHeight))
			if pickupstuff == nil {
				// 区块高度不匹配什么都不做
				continue
			}
			newstuff, newhx := pickupstuff.CalculateBlockHashByMiningResult(&result, true)
			var isMintSuccess = false // 是否真的挖掘成功
			if result.MintSuccessed.Check() {
				// 挖掘成功，开始验证
				// 判断哈希满足要求
				newblock := newstuff.GetHeadMetaBlock()
				if difficulty.CheckHashDifficultySatisfyByBlock(newhx, newblock) {
					// 满足难度 上报
					if m.server.service_tcp != nil {
						message.MsgSendToTcpConn(m.server.service_tcp, message.MinerWorkMsgTypeReportMiningResult, result.Serialize())
					}
					isMintSuccess = true
				} else {
					// 不满足难度， 什么都不做
					fmt.Println("relay service 不满足难度， 什么都不做")
					diffhash := difficulty.Uint32ToHash(newblock.GetHeight(), newblock.GetDifficulty())
					diffhex := hex.EncodeToString(diffhash)
					fmt.Println(newblock.GetHeight(), newhx.ToHex(), diffhex, hex.EncodeToString(newblock.GetNonceByte()), newblock.GetNonceByte())
					fmt.Println(hex.EncodeToString(blocks.CalculateBlockHashBaseStuff(newblock)))
				}
				// 处理完毕
			}

			// 没有挖掘成功的话，忽略此消息 或 写入算力统计
			if m.server.config.IsAcceptHashrate == false {
				// 不接受算力统计
				continue
			}
			// 上报算力统计
			if m.server.config.IsReportHashrate == true {
				// 上报
				if m.server.service_tcp != nil {
					message.MsgSendToTcpConn(m.server.service_tcp, message.MinerWorkMsgTypeReportMiningResult, result.Serialize())
				}
			}
			// 写入算力统计
			m.server.saveMiningResultToStore(m.rwdaddr, isMintSuccess, newstuff)

		} else {
			return fmt.Errorf("Not supported msg type")
		}
	}
}
