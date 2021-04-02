package minerworker

import (
	"fmt"
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
		m.powWorker.StopAllMining() // 关闭挖矿
		m.conn = nil                // 表示断开连接
		//fmt.Println("handleConn end", m.conn)
	}()

	// 已连接上
	respmsgobj, err := message.HandleConnectToServer(conn, &m.config.Rewards)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 是否接受算力统计
	if respmsgobj.AcceptHashrateStatistics.Is(false) {
		m.config.IsReportHashrate = false // 不接受统计
		m.powWorker.CloseUploadHashrate() // 关闭统计
		fmt.Print(" (note: pool is not accept PoW power statistics) ")
	}

	firstshowconnectok := true

	// 循环收取挖矿消息
	for {

		//fmt.Println("循环收取挖矿消息")
		msgty, msgbody, err := message.MsgReadFromTcpConn(conn, 0)
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				// 服务器关闭
				fmt.Println("\n[Miner Worker] WARNING: Server close the tcp connect, reconnection will be initiated in two minutes...")
			} else {
				fmt.Println(err)
			}
			break
		}

		if msgty == message.MinerWorkMsgTypeMiningBlock {
			var stuff = &message.MsgPendingMiningBlockStuff{}
			_, err := stuff.Parse(msgbody, 0)
			if err != nil {
				fmt.Println("message.MsgPendingMiningBlockStuff.Parse Error", err)
				continue
			}
			m.pendingMiningBlockStuff = stuff // 挖矿 stuff

			if firstshowconnectok {
				firstshowconnectok = false
				fmt.Println("connected successfully.")
			}

			// 执行下一个挖矿
			go m.powWorker.DoNextMining(stuff.BlockHeadMeta.GetHeight())

		} else {
			fmt.Printf("message type [%d] not supported\n", msgty)
			continue
		}
	}

}
