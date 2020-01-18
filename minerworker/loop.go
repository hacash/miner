package minerworker

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/miner/message"
	"time"
)

func (p *MinerWorker) loop() {

	sendPingMsgToPoolServer := time.NewTicker(time.Second * 35)
	checkPongMsgReturn := time.NewTicker(time.Second * 4)
	restartTick := time.NewTicker(time.Second * 13)
	//notEndSuccessMsg := time.NewTicker(time.Minute * 3)

	for {
		select {

		/*
		case <-notEndSuccessMsg.C:
			if p.currentMiningStatusSuccess {
				p.currentMiningStatusSuccess = false
				if p.client != nil {
					p.client.conn.Close() // restart next mining
				}
			}
		 */

		case <-	sendPingMsgToPoolServer.C:
			if p.client != nil && p.client.workBlockHeight > 0 {
				pingmsg := []byte("ping")
				tarhei := fields.VarInt5(p.client.workBlockHeight)
				heibts, _ := tarhei.Serialize()
				p.client.conn.Write( append(pingmsg, heibts...) )
				ctime := time.Now()
				p.client.pingtime = &ctime
				//fmt.Println("send ping", p.client)
			}

		case <- checkPongMsgReturn.C:
			//fmt.Print("chenk pong... ", p.client)
			if p.client != nil && p.client.pingtime != nil {
				if p.client.pingtime.Add(time.Second * time.Duration(5)).Before(time.Now()) {
					p.client.conn.Close() // force close with no pong
					fmt.Println(" --[ force close with no pong ]-- ")
				}else{
					//fmt.Println("ok")
				}
			}

		case msg := <-p.miningOutputCh:

			//fmt.Println( "msg := <- p.miningOutputCh:")
			//fmt.Println("msg: ", msg.BlockHeadMeta.GetHeight(), msg.CoinbaseMsgNum, msg.Status, msg.NonceBytes, msg)
			p.statusMutex.Lock()

			client := p.pickTargetClient( msg.BlockHeadMeta.GetHeight() )
			//fmt.Println("pickTargetClient", client)
			if client != nil {

				msg.BlockHeadMeta.SetNonce(binary.BigEndian.Uint32(msg.NonceBytes))
				msg.BlockHeadMeta.Fresh()

				if msg.Status == message.PowMasterMsgStatusSuccess || msg.Status == message.PowMasterMsgStatusMostPowerHash {
					msgbytes, _ := msg.Serialize()
					client.conn.Write(msgbytes) // send success
				}
				if msg.Status == message.PowMasterMsgStatusSuccess {
					//p.currentMiningStatusSuccess = true // set mining status
					fmt.Print("OK.\n== ⬤ == Successfully mining block height: ", msg.BlockHeadMeta.GetHeight(), ", hash: ", msg.BlockHeadMeta.Hash().ToHex(), "\n")
				}
				if msg.Status == message.PowMasterMsgStatusMostPowerHash {
					fmt.Print("upload power hash: ", hex.EncodeToString(msg.BlockHeadMeta.Hash()[0:12]), "... ok.\n")
					/*if p.client != nil {
						p.client.conn.Close() // next mining
					}*/
				}
				if client.setend {
					client.conn.Close() // close
				}
				client.setend = true

			}

			p.statusMutex.Unlock()


		case <-p.immediateStartConnectCh:
			err := p.startConnect()
			if err != nil {
				fmt.Println(err)
			}

		case <-restartTick.C:
			if p.client == nil {
				p.immediateStartConnectCh <- true
			}

		}
	}

}
