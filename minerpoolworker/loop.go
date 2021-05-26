package minerpoolworker

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/miner/message"
	"github.com/hacash/mint/difficulty"
	"math/big"
	"time"
)

func (p *MinerPoolWorker) loop() {

	sendPingMsgToPoolServer := time.NewTicker(time.Second * 55)
	checkPongMsgReturn := time.NewTicker(time.Second * 10)
	restartTick := time.NewTicker(time.Second * 13)
	//notEndSuccessMsg := time.NewTicker(time.Minute * 3)

	for {
		select {

		/*
			case <-notEndSuccessMsg.C:
				if p.currentMiningStatusSuccess {
					p.currentMiningStatusSuccess = false
					if p.client != nil {
						p.client.conn.notifyClose() // restart next mining
					}
				}
		*/

		case <-sendPingMsgToPoolServer.C:
			if p.client != nil && p.client.workBlockHeight > 0 {
				pingmsg := []byte("ping")
				tarhei := fields.VarUint5(p.client.workBlockHeight)
				heibts, _ := tarhei.Serialize()
				ctime := time.Now()
				p.client.pingtime = &ctime
				p.client.conn.Write(append(pingmsg, heibts...))
				//fmt.Println("send ping", p.client)
			}

		case <-checkPongMsgReturn.C:
			//fmt.Print("chenk pong... ", p.client)
			if p.client != nil && p.client.pingtime != nil {
				if p.client.pingtime.Add(time.Second * time.Duration(21)).Before(time.Now()) {
					p.client.conn.Close() // force close with no pong
					fmt.Println(" --[ force close with no pong ]-- ")
				} else {
					//fmt.Println("ok")
				}
			}

		case msg := <-p.miningOutputCh:

			//fmt.Println( "msg := <- p.miningOutputCh:")
			//fmt.Println("msg: ", msg.BlockHeadMeta.GetHeight(), msg.CoinbaseMsgNum, msg.Status, msg.NonceBytes, msg)
			p.statusMutex.Lock()

			block_height := msg.BlockHeadMeta.GetHeight()

			client := p.pickTargetClient(block_height)
			//fmt.Println("pickTargetClient", client)
			if client != nil {

				msg.BlockHeadMeta.SetNonce(binary.BigEndian.Uint32(msg.NonceBytes))
				msg.BlockHeadMeta.Fresh()

				var hashrateshow string = "-"
				var usetimesec int64 = 0
				var block_hash fields.Hash = nil

				if msg.Status == message.PowMasterMsgStatusSuccess ||
					msg.Status == message.PowMasterMsgStatusMostPowerHash ||
					msg.Status == message.PowMasterMsgStatusMostPowerHashAndRequestNextMining {
					msgbytes, _ := msg.Serialize()
					go client.conn.Write(msgbytes) // send success
					// power worth
					block_hash = msg.BlockHeadMeta.Hash()
					hxworth := difficulty.CalculateHashWorth(msg.BlockHeadMeta.GetHeight(), block_hash)
					usetimesec = int64(time.Now().Sub(client.miningStartTime).Seconds())
					if usetimesec == 0 {
						usetimesec = 1
					}
					//fmt.Println( usetimesec )
					hashrate := new(big.Int).Div(hxworth, big.NewInt(usetimesec))
					//hashrateshow = difficulty.ConvertPowPowerToShowFormat(hashrate)
					//hashrateshow += ", " + p.addPowerLogReturnShow(hashrate)
					hashrateshow = p.addPowerLogReturnShow(hashrate)
				}
				if msg.Status == message.PowMasterMsgStatusSuccess {
					//p.currentMiningStatusSuccess = true // set mining status
					fmt.Printf("OK.\n[⬤◆◆] Successfully mined a block height: %d, hash: %s, time: %ds, hashrate: %s. \n", block_height, block_hash.ToHex(), usetimesec, hashrateshow)
				}
				if msg.Status == message.PowMasterMsgStatusMostPowerHash || msg.Status == message.PowMasterMsgStatusMostPowerHashAndRequestNextMining {
					fmt.Printf("upload hash: %d, %s..., time: %ds, hashrate: %s ok.\n", block_height, hex.EncodeToString(block_hash[0:12]), usetimesec, hashrateshow)
					/*if p.client != nil {
						p.client.conn.notifyClose() // next mining
					}*/
				}
				if client.setend {
					client.conn.Close() // close
				}
				client.setend = true

			}

			p.statusMutex.Unlock()

		case <-p.immediateStartConnectCh:
			if p.isInConnecting == false {
				err := p.startConnect()
				if err != nil {
					fmt.Println(err)
				}
			}

		case <-restartTick.C:
			if p.client == nil && p.isInConnecting == false {
				p.immediateStartConnectCh <- true
			}

		}
	}

}
