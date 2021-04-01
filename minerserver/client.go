package minerserver

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/miner/message"
	"github.com/hacash/mint/difficulty"
	"math/rand"
	"net"
)

type MinerServerClinet struct {
	server *MinerServer
	id     uint64
	conn   *net.TCPConn
}

func NewMinerServerClinet(server *MinerServer, conn *net.TCPConn) *MinerServerClinet {
	cid := rand.Uint64()
	return &MinerServerClinet{
		server: server,
		id:     cid,
		conn:   conn,
	}
}

// Handle
func (m *MinerServerClinet) Handle() error {
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
				continue // 没有挖掘成功的话，忽略此消息
			}

			// 挖掘完成，开始验证
			newstuff, newhx := m.server.penddingBlockMsg.CalculateBlockHashByMiningResult(&result, true)
			// 判断哈希满足要求
			newblock := newstuff.GetHeadMetaBlock()
			if difficulty.CheckHashDifficultySatisfyByBlock(newhx, newblock) {
				// 满足难度 写入区块链
				//fmt.Println( "GetTransactionCount:", newblock.GetTransactionCount(), )
				newblock.SetOriginMark("mining") // set origin
				m.server.successMintCh <- newblock
			} else {
				// 不满足难度， 什么都不做

				fmt.Println("不满足难度， 什么都不做")
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
