package minerworker

import (
	"encoding/binary"
	"fmt"
	"github.com/hacash/miner/message"
	"time"
)

func (p *MinerWorker) loop() {

	restartTick := time.NewTicker(time.Second * 7)

	for {
		select {
		case msg := <-p.miningOutputCh:

			//fmt.Println( "msg := <- p.miningOutputCh:")
			//fmt.Println("msg: ", msg.CoinbaseMsgNum, msg.Status, msg.NonceBytes, msg)

			msg.BlockHeadMeta.SetNonce(binary.BigEndian.Uint32(msg.NonceBytes))
			msg.BlockHeadMeta.Fresh()

			if msg.Status == message.PowMasterMsgStatusSuccess || msg.Status == message.PowMasterMsgStatusMostPowerHash {
				msgbytes, _ := msg.Serialize()
				if p.conn != nil {
					p.conn.Write(msgbytes) // send success
				}
			}
			if msg.Status == message.PowMasterMsgStatusSuccess {
				fmt.Print("\n== ⬤ == Successfully mining block height: ", msg.BlockHeadMeta.GetHeight(), ", hash: ", msg.BlockHeadMeta.Hash().ToHex(), ", rewards: ", p.config.Rewards.ToReadable())
			}
			if msg.Status == message.PowMasterMsgStatusMostPowerHash {
				fmt.Print("upload power hash:", msg.BlockHeadMeta.Hash().ToHex())
				if p.conn != nil {
					p.conn.Close() // next mining
				}
			}

			fmt.Println("")

		case <-p.immediateStartConnectCh:
			err := p.startConnect()
			if err != nil {
				fmt.Println(err)
			}

		case <-restartTick.C:
			if p.conn == nil {
				p.immediateStartConnectCh <- true
			}

		}
	}

}
