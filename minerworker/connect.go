package minerworker

import (
	"bytes"
	"fmt"
	"github.com/hacash/miner/message"
	"net"
	"os"
	"time"
)

const (
	MsgMarkNotReadyYet      = "not_ready_yet"
	MsgMarkTooMuchConnect   = "too_many_connect"
	MsgMarkEndCurrentMining = "end_current_mining"
	MsgMarkPong             = "pong"

)

func (p *MinerWorker) startConnect() error {

	conn, err := net.DialTCP("tcp", nil, p.config.PoolAddress)
	if err != nil {
		return err
	}

	fmt.Print("connecting... ")

	go p.handleConn(conn)

	return nil

}

func (p *MinerWorker) handleConn(conn *net.TCPConn) {
	client := NewClient(conn)
	p.client = client
	// send reward address
	//fmt.Println([]byte(p.config.Rewards))
	conn.Write(p.config.Rewards)

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

		databuf.Write( segdata[0:rn] )
		data := databuf.Bytes()
		rn = len(data)

		//fmt.Println("MinerWorker: rn, err := conn.Read(segdata)", message.PowMasterMsgSize, len(segdata[0:rn]), segdata[0:rn])
		//fmt.Println("MinerWorker: rn, err := conn.Read(segdata)", string(segdata[0:rn]))

		if rn == len(MsgMarkPong) && bytes.Compare([]byte(MsgMarkPong), data) == 0 {

			client.pingtime = nil // reset ping time

		}else if rn == len(MsgMarkTooMuchConnect) && bytes.Compare([]byte(MsgMarkTooMuchConnect), data) == 0 {
			// wait for min
			fmt.Println("pool return: " + MsgMarkTooMuchConnect)
			fmt.Println("There are too many ore pool connections. The connection has been refused. Please contact your ore pool service provider.")
			fmt.Println("矿池连接数太多，已拒绝连接，请联系您的矿池服务商。")
			os.Exit(0)

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
			if client.setend {
				client.conn.Close() // close
			}else{
				fmt.Print("ending... ")
				client.setend = true
			}
			p.statusMutex.Unlock()


		} else if rn == message.PowMasterMsgSize {

			p.worker.StopMining()
			// start mining
			powmsg := message.NewPowMasterMsg()
			powmsg.Parse(data, 0)

			client.workBlockHeight = powmsg.BlockHeadMeta.GetHeight()
			p.clients.Add(client)

			//fmt.Println("Excavate",  powmsg.CoinbaseMsgNum, powmsg.BlockHeadMeta)
			fmt.Print("do mining height:‹", powmsg.BlockHeadMeta.GetHeight(), "›, cbmn:", powmsg.CoinbaseMsgNum, "... ")
			// do work
			p.worker.SetCoinbaseMsgNum(uint32(powmsg.CoinbaseMsgNum))
			//time.Sleep(time.Second)
			p.worker.Excavate(powmsg.BlockHeadMeta, p.miningOutputCh)


		} else {

			goto READNEXTDATASEG

		}
	}

	//fmt.Println( "------ - --- - - -- break conn.Close()" )

	conn.Close()

	p.worker.StopMining()

	p.client = nil

	p.immediateStartConnectCh <- true
}
