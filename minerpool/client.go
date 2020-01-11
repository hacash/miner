package minerpool

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/miner/message"
	"github.com/hacash/mint/difficulty"
	"math/big"
	"net"
)

type Client struct {
	belongAccount *Account

	conn *net.TCPConn

	address fields.Address

	workBlock interfaces.Block

	coinbaseMsgNum uint32 // > 0

}

func NewClient(acc *Account, conn *net.TCPConn, workBlock interfaces.Block) *Client {
	return &Client{
		belongAccount:  acc,
		conn:           conn,
		workBlock:      workBlock,
		address:        nil,
		coinbaseMsgNum: 0,
	}
}

// 上报挖矿结果

func (c *Client) postPowResult(msg *message.PowMasterMsg) {
	block := msg.BlockHeadMeta

	if c.workBlock.GetHeight() != block.GetHeight() {
		return // error
	}

	block.SetNonce(binary.BigEndian.Uint32(msg.NonceBytes))
	block.Fresh()
	blkhash := block.HashFresh()

	//fmt.Println("postPowResult", uint32(msg.CoinbaseMsgNum) )

	//fmt.Println( "  -  1  -   postPowResult(msg *message.PowMasterMsg)" )

	minerpool := c.belongAccount.realtimePeriod.minerpool

	// 添加算力统计
	if msg.Status != message.PowMasterMsgStatusSuccess {
		c.belongAccount.addPowWorth(blkhash)
		return
	}

	// 挖出区块
	if msg.Status == message.PowMasterMsgStatusSuccess {
		targetdiffhash := difficulty.Uint32ToHash(block.GetHeight(), block.GetDifficulty())

		//fmt.Println("targetdiffhash", hex.EncodeToString(targetdiffhash))
		//fmt.Println("blkhash", blkhash.ToHex())

		targetbig := new(big.Int).SetBytes(targetdiffhash)
		blkbig := new(big.Int).SetBytes(blkhash)
		if blkbig.Cmp(targetbig) == 1 {
			fmt.Println("fail mining pool pow worker result check: need %s but got %s", hex.EncodeToString(targetdiffhash), hex.EncodeToString(blkhash))
			c.conn.Close() // 关闭连接
			return
		}
		// success find block
		go func() {
			minerpool.successFindBlockCh <- &findBlockMsg{
				msg:     msg,
				account: c.belongAccount,
			}
		}()
		return
	}
}
