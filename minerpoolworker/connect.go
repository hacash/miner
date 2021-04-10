package minerpoolworker

import (
	"bytes"
	"fmt"
	"github.com/hacash/miner/message"
	"net"
	"time"
)

const (
	MsgMarkNotReadyYet      = "not_ready_yet"
	MsgMarkTooMuchConnect   = "too_many_connect"
	MsgMarkEndCurrentMining = "end_current_mining"
	MsgMarkPong             = "pong"
)

func (p *MinerWorker) startConnect() error {

	p.isInConnecting = true

	conn, err := net.DialTCP("tcp", nil, p.config.PoolAddress)
	if err != nil {
		p.isInConnecting = false
		return err
	}

	go p.handleConn(conn)

	return nil

}

func (p *MinerWorker) handleConn(conn *net.TCPConn) {

	fmt.Print("connecting miner pool... ")

	// send reward address
	//fmt.Println([]byte(p.config.Rewards))
	_, e := conn.Write(p.config.Rewards)
	if e != nil {
		fmt.Println("----------[ERROR]----------")
		fmt.Println("Cannot connect to", conn.RemoteAddr().String())
		fmt.Println("----------[ERROR]----------")
		return
	}

	fmt.Println("ok.")

	// read msg
	segdata := make([]byte, 1024)

	for {

		databuf := bytes.NewBuffer([]byte{})

	READNEXTDATASEG:

		//fmt.Println("READNEXTDATASEG")

		rn, err := conn.Read(segdata)
		if err != nil {
			//fmt.Println(err)
			break
		}

		//fmt.Println(segdata[0:rn])

		databuf.Write(segdata[0:rn])
		data := databuf.Bytes()
		rn = len(data)

		//fmt.Println("MinerWorker: rn, err := conn.Read(segdata)", message.PowMasterMsgSize, len(segdata[0:rn]), segdata[0:rn])
		//fmt.Println("MinerWorker: rn, err := conn.Read(segdata)", string(segdata[0:rn]))

		if rn == len(MsgMarkPong) && bytes.Compare([]byte(MsgMarkPong), data) == 0 {

			if p.client != nil {
				p.client.pingtime = nil // reset ping time
			}

		} else if rn == len(MsgMarkTooMuchConnect) && bytes.Compare([]byte(MsgMarkTooMuchConnect), data) == 0 {
			// wait for min
			fmt.Println("pool return: " + MsgMarkTooMuchConnect)
			fmt.Println("There are too many ore pool connections. The connection has been refused. Please contact your ore pool service provider.")
			fmt.Println("矿池连接数太多，已拒绝连接，请联系您的矿池服务商。")
			time.Sleep(time.Second * 30)
			break

		} else if rn == len(MsgMarkNotReadyYet) && bytes.Compare([]byte(MsgMarkNotReadyYet), data) == 0 {
			// wait for min
			fmt.Println("pool return: " + MsgMarkNotReadyYet)
			time.Sleep(time.Second * 5)
			break

		} else if rn == len(MsgMarkEndCurrentMining) && bytes.Compare([]byte(MsgMarkEndCurrentMining), data) == 0 {

			p.statusMutex.Lock()
			//fmt.Println( "  -  1  -  p.worker.StopMining() ", p.currentMiningStatusSuccess )
			// 结束挖矿，等待上报挖矿结果
			p.worker.StopMining()
			if p.client != nil {
				if p.client.setend {
					p.client.conn.Close() // close
				} else {
					fmt.Print("next... ")
					p.client.setend = true
				}
			}
			p.statusMutex.Unlock()

		} else if rn == message.PowMasterMsgSize {

			// start mining
			powmsg := message.NewPowMasterMsg()
			_, e := powmsg.Parse(data, 0)
			if e != nil {
				// 解析消息发生错误，什么也不做
				continue
			}
			tarBlockHeight := powmsg.BlockHeadMeta.GetHeight()

			if p.currentPowMasterMsg != nil && p.client != nil && p.isInConnecting &&
				p.currentPowMasterMsg.BlockHeadMeta.GetHeight() == tarBlockHeight &&
				p.currentPowMasterCreateTime.Add(time.Second*3).After(time.Now()) {
				//p.currentPowMasterMsg.CoinbaseMsgNum == powmsg.CoinbaseMsgNum {
				// 5秒内重复挖矿消息，忽略本次消息
				//fmt.Print(" -ignore duplicate mining messages- ")
				fmt.Print("idmm... ")
			} else {
				// 执行挖矿
				p.currentPowMasterMsg = powmsg
				p.currentPowMasterCreateTime = time.Now()

				client := NewClient(conn)
				client.workBlockHeight = tarBlockHeight
				p.clients[client.workBlockHeight] = client
				p.client = client

				// stop prev mining
				p.worker.StopMining()

				//fmt.Println("Excavate",  powmsg.CoinbaseMsgNum, powmsg.BlockHeadMeta)
				fmt.Print("do mining height:‹", tarBlockHeight, "›, cbmn:", powmsg.CoinbaseMsgNum, "... ")
				// do work
				p.worker.SetCoinbaseMsgNum(uint32(powmsg.CoinbaseMsgNum))
				//time.Sleep(time.Second)
				p.worker.Excavate(powmsg.BlockHeadMeta, p.miningOutputCh)

			}

		} else {

			goto READNEXTDATASEG

		}
	}

	//fmt.Println( "------ - --- - - -- break conn.notifyClose()" )

	conn.Close()

	p.worker.StopMining()

	p.client = nil
	p.isInConnecting = false

	p.immediateStartConnectCh <- true
}
