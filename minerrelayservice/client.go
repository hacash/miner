package minerrelayservice

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/miner/message"
	"github.com/hacash/mint/difficulty"
	"math/rand"
	"net"
)

type ConnClient struct {
	server *RelayService
	id     uint64
	conn   *net.TCPConn
}

func NewConnClient(server *RelayService, conn *net.TCPConn) *ConnClient {
	cid := rand.Uint64()
	return &ConnClient{
		server: server,
		id:     cid,
		conn:   conn,
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
			if result.MintSuccessed != 1 {
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
					continue
				}
				// 自己写入算力统计

			}
			// 挖掘成功，开始验证
			newstuff, newhx := m.server.penddingBlockStuff.CalculateBlockHashByMiningResult(&result, true)
			// 判断哈希满足要求
			newblock := newstuff.GetHeadMetaBlock()
			if difficulty.CheckHashDifficultySatisfyByBlock(newhx, newblock) {
				// 满足难度 上报
				if m.server.service_tcp != nil {
					message.MsgSendToTcpConn(m.server.service_tcp, message.MinerWorkMsgTypeReportMiningResult, result.Serialize())
				}
			} else {
				// 不满足难度， 什么都不做

				fmt.Println("relay service 不满足难度， 什么都不做")
				diffhash := difficulty.Uint32ToHash(newblock.GetHeight(), newblock.GetDifficulty())
				diffhex := hex.EncodeToString(diffhash)
				fmt.Println(newblock.GetHeight(), newhx.ToHex(), diffhex, hex.EncodeToString(newblock.GetNonceByte()), newblock.GetNonceByte())
				fmt.Println(hex.EncodeToString(blocks.CalculateBlockHashBaseStuff(newblock)))
			}

		} else {
			return fmt.Errorf("Not supported msg type")
		}
	}
}
