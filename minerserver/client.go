package minerserver

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/miner/interfaces"
	"github.com/hacash/miner/message"
	"github.com/hacash/mint/difficulty"
	"math/rand"
	"net"
)

type MinerServerClient struct {
	server *MinerServer
	id     uint64
	conn   *net.TCPConn
}

func NewMinerServerClinet(server *MinerServer, conn *net.TCPConn) *MinerServerClient {
	cid := rand.Uint64()
	return &MinerServerClient{
		server: server,
		id:     cid,
		conn:   conn,
	}
}

// Handle
func (m *MinerServerClient) Handle() error {
	for {
		// Read message
		msgty, msgbody, err := message.MsgReadFromTcpConn(m.conn, 0)
		if err != nil {
			return err
		}
		//fmt.Println(msgty, msgbody)
		// Parse message
		if msgty == message.MinerWorkMsgTypeReportMiningResult {

			var result = interfaces.PoWResultShortData{}
			_, err := result.Parse(msgbody, 0)
			if err != nil {
				return err
			}
			if m.server.penddingBlockMsg == nil {
				fmt.Printf(" m.server.penddingBlockMsg == nil block <%d> !!!! continue\n",
					result.BlockHeight)
				continue
			}
			// handle
			if !result.FindSuccess.Check() {
				fmt.Printf("!result.FindSuccess.Check() !!!! block <%d> continue\n",
					result.BlockHeight)
				continue // If mining is not successful, ignore this message
			}

			// Excavation completed, start verification
			newhx, newblock, err := m.server.penddingBlockMsg.CalculateBlockHashByMiningResult(&result, true)
			//fmt.Println("get:::::", result.BlockNonce, result.CoinbaseNonce.ToHex(),
			//	newblock.GetHeight(),
			//	newblock.GetMrklRoot().ToHex())
			if err != nil {
				fmt.Printf("block %d continue \n", result.BlockHeight)
				fmt.Println("m.server.penddingBlockMsg.CalculateBlockHashByMiningResult ERROR: ", err.Error())
				//return err
				// not match penddingBlock , do nothing
				continue
			}
			// Judge whether the hash meets the requirements
			if !difficulty.CheckHashDifficultySatisfyByBlock(newhx, newblock) {

				// 不满足难度， 什么都不做
				fmt.Println("不满足难度， 什么都不做")
				diffhash := difficulty.Uint32ToHash(newblock.GetHeight(), newblock.GetDifficulty())
				diffhex := hex.EncodeToString(diffhash)
				fmt.Println(newblock.GetHeight(), newhx.ToHex(), diffhex, hex.EncodeToString(newblock.GetNonceByte()), newblock.GetNonceByte())
				//fmt.Println(hex.EncodeToString(blocks.CalculateBlockHashBaseStuff(newblock)))
				continue
			}
			// FIND SUCCESS !!!!!!!!
			fmt.Printf("FIND SUCCESS !!!!!!!! block <%d> m.server.successMintCh <- newblock\n", result.BlockHeight)
			// Writing blockchain to meet the difficulty
			//fmt.Println( "GetTransactionCount:", newblock.GetTransactionCount(), )
			newblock.SetOriginMark("mining") // set origin
			m.server.successMintCh <- newblock

		} else {
			//fmt.Println("Not supported msg type, close the client conn!")
			return fmt.Errorf("Not supported msg type")
		}
	}
}
