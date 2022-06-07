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
		// Read message
		msgty, msgbody, err := message.MsgReadFromTcpConn(m.conn, 0)
		if err != nil {
			return err
		}
		// Parse message
		if msgty == message.MinerWorkMsgTypeReportMiningResult {
			var result = message.MsgReportMiningResult{}
			_, err := result.Parse(msgbody, 0)
			if err != nil {
				return err
			}
			// handle
			if result.MintSuccessed != 1 {
				continue // If mining is not successful, ignore this message
			}

			// Excavation completed, start verification
			newstuff, newhx := m.server.penddingBlockMsg.CalculateBlockHashByMiningResult(&result, true)
			// Judge whether the hash meets the requirements
			newblock := newstuff.GetHeadMetaBlock()
			if difficulty.CheckHashDifficultySatisfyByBlock(newhx, newblock) {
				// Writing blockchain to meet the difficulty
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
