package minerworker

import (
	"fmt"
	interfaces2 "github.com/hacash/miner/interfaces"
	"github.com/hacash/miner/message"
	"net"
	"strings"
)

func (p *MinerWorker) startConnect() error {

	fmt.Print("connecting miner server...")

	conn, err := net.DialTCP("tcp", nil, p.config.PoolAddress)
	if err != nil {
		return err
	}

	go p.handleConn(conn)

	return nil

}

func (m *MinerWorker) handleConn(conn *net.TCPConn) {

	m.conn = conn

	//fmt.Println("handleConn start", m.conn)
	defer func() {
		m.powWorker.StopMining() // Close mining
		m.conn = nil             // Indicates disconnection
		//fmt.Println("handleConn end", m.conn)
	}()

	// Connected
	respmsgobj, err := message.HandleConnectToServer(conn, &m.config.Rewards)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Whether to accept calculation force statistics
	if respmsgobj.AcceptHashrateStatistics.Is(false) {
		m.config.IsReportHashrate = false // Statistics not accepted
		//m.powMaster.CloseUploadHashrate() // Turn off statistics
		fmt.Print(" (note: pool is not accept PoW power statistics) ")
	}

	firstshowconnectok := true

	// Collect mining messages circularly
	for {

		//fmt.Println("循环收取挖矿消息")
		msgty, msgbody, err := message.MsgReadFromTcpConn(conn, 0)
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				// Server shutdown
				fmt.Println("\n[Miner Worker] WARNING: Server close the tcp connect, reconnection will be initiated in two minutes...")
			} else {
				fmt.Println(err)
			}
			break
		}

		if msgty == message.MinerWorkMsgTypeMiningBlock {

			var stuff = &interfaces2.PoWStuffOverallData{}
			_, err := stuff.Parse(msgbody, 0)
			if err != nil {
				fmt.Println("message.MsgPendingMiningBlockStuff.Parse Error", err)
				continue
			}
			//fmt.Println(hex.EncodeToString(msgbody), stuff.CoinbaseTx.Address.ToReadable(), stuff.CoinbaseTx.ExtendDataVersion )
			//m.pendingMiningBlockStuff = stuff // Mining stuff

			if firstshowconnectok {
				firstshowconnectok = false
				fmt.Println("connected successfully.")
			}

			// Perform next mining
			go m.Excavate(stuff)

		} else {
			fmt.Printf("message type [%d] not supported\n", msgty)
			continue
		}
	}

}

func (m *MinerWorker) Excavate(stuff *interfaces2.PoWStuffOverallData) {
	m.powWorker.StopMining()

	var result, err = m.powWorker.DoMining(stuff)
	if err != nil {
		fmt.Println(err)
		return
	}
	if result == nil {
		return
	}
	if !result.FindSuccess.Check() && !m.config.IsReportHashrate {
		return
	}
	// upload hashrate
	upretdts, err := result.GetShortData().Serialize()
	if err != nil {
		fmt.Println(err)
		return
	}

	// report to server
	go message.MsgSendToTcpConn(m.conn, message.MinerWorkMsgTypeReportMiningResult, upretdts)
	//fmt.Println(hex.EncodeToString(upretdts))
}
