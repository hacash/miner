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
	rwdaddr fields.Address // Reward address
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
			pickupstuff := m.server.checkoutMiningStuff(uint64(result.BlockHeight))
			if pickupstuff == nil {
				// Block height mismatch do nothing
				continue
			}
			newstuff, newhx := pickupstuff.CalculateBlockHashByMiningResult(&result, true)
			var isMintSuccess = false // Whether the mining is really successful
			if result.MintSuccessed.Check() {
				// Mining succeeded, start verification
				// Judge whether the hash meets the requirements
				newblock := newstuff.GetHeadMetaBlock()
				if difficulty.CheckHashDifficultySatisfyByBlock(newhx, newblock) {
					// Meet the difficulty Report
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
				// Processing completed
			}

			// If the mining is not successful, ignore this message or write the calculation force statistics
			if m.server.config.IsAcceptHashrate == false {
				// Calculation force statistics are not acceptable
				continue
			}
			// Submit calculation force statistics
			if m.server.config.IsReportHashrate == true {
				// Escalation
				if m.server.service_tcp != nil {
					message.MsgSendToTcpConn(m.server.service_tcp, message.MinerWorkMsgTypeReportMiningResult, result.Serialize())
				}
			}
			// Write calculation force statistics
			go m.server.saveMiningResultToStore(m.rwdaddr, isMintSuccess, newstuff)

		} else {
			return fmt.Errorf("Not supported msg type")
		}
	}
}
