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
	MsgMarkEndCurrentMining = "end_current_mining"
	MsgMarkTooMuchConnect   = "too_many_connect"
)

func (p *MinerWorker) startConnect() error {

	conn, err := net.DialTCP("tcp", nil, p.config.PoolAddress)
	if err != nil {
		return err
	}

	fmt.Print("connect to pool ", p.config.PoolAddress.String(), " ... ")

	go p.handleConn(conn)

	return nil

}

func (p *MinerWorker) handleConn(conn *net.TCPConn) {
	p.conn = conn

	// send reward address
	conn.Write(p.config.Rewards)

	// read msg
	segdata := make([]byte, 1024)

	for {

		rn, err := conn.Read(segdata)
		if err != nil {
			// fmt.Println(err)
			break
		}

		//fmt.Println("MinerWorker: rn, err := conn.Read(segdata)", message.PowMasterMsgSize, len(segdata[0:rn]), segdata[0:rn])
		//fmt.Println("MinerWorker: rn, err := conn.Read(segdata)", string(segdata[0:rn]))

		if rn == len(MsgMarkTooMuchConnect) && bytes.Compare([]byte(MsgMarkTooMuchConnect), segdata[0:rn]) == 0 {
			// wait for min
			fmt.Println("pool return: " + MsgMarkTooMuchConnect)
			fmt.Println("矿池连接数太多，已拒绝连接，请联系您的矿池服务商。")
			os.Exit(0)

		} else if rn == len(MsgMarkNotReadyYet) && bytes.Compare([]byte(MsgMarkNotReadyYet), segdata[0:rn]) == 0 {
			// wait for min
			fmt.Println("pool return: " + MsgMarkNotReadyYet)
			time.Sleep(time.Second * 5)
			break

		} else if rn == len(MsgMarkEndCurrentMining) && bytes.Compare([]byte(MsgMarkEndCurrentMining), segdata[0:rn]) == 0 {

			//fmt.Println( "  -  1  -  p.worker.StopMining() " )

			// 结束挖矿，上报挖矿结果
			p.worker.StopMining()

		} else if rn == message.PowMasterMsgSize {
			// start mining
			powmsg := message.NewPowMasterMsg()
			powmsg.Parse(segdata[0:rn], 0)
			//fmt.Println("Excavate",  powmsg.CoinbaseMsgNum, powmsg.BlockHeadMeta)
			fmt.Print("mining block height: ", powmsg.BlockHeadMeta.GetHeight(), ", cbn: ", powmsg.CoinbaseMsgNum, " ... ")
			// do work
			p.worker.SetCoinbaseMsgNum(uint32(powmsg.CoinbaseMsgNum))
			p.worker.Excavate(powmsg.BlockHeadMeta, p.miningOutputCh)
		}
	}

	conn.Close()

	p.conn = nil

	p.immediateStartConnectCh <- true
}
