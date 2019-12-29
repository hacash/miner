package minerworker

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/hacash/miner/message"
	"time"
)

func (p *MinerWorker) loop() {

	restartTick := time.NewTicker(time.Second * 13)
	notEndSuccessMsg := time.NewTicker(time.Minute * 3)

	for {
		select {

		case <-notEndSuccessMsg.C:
			if p.currentMiningStatusSuccess {
				p.currentMiningStatusSuccess = false
				if p.conn != nil {
					p.conn.Close() // restart next mining
				}
			}

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
				p.currentMiningStatusSuccess = true // set mining status
				fmt.Print("OK.\n\n== ⬤ == Successfully mining block height: ", msg.BlockHeadMeta.GetHeight(), ", hash: ", msg.BlockHeadMeta.Hash().ToHex(), ", rewards: ", p.config.Rewards.ToReadable(), "\n")
			}
			if msg.Status == message.PowMasterMsgStatusMostPowerHash {
				fmt.Print("upload power hash: ", hex.EncodeToString(msg.BlockHeadMeta.Hash()[0:12]), " ok.")
				if p.conn != nil {
					p.conn.Close() // next mining
				}
			}

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
