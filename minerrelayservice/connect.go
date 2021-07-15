package minerrelayservice

import (
	"fmt"
	"github.com/hacash/miner/message"
	"net"
	"strings"
	"time"
)

func (r *RelayService) connectToService() {

	fmt.Printf("connecting server <%s>... ", r.config.ServerAddress.String())

	conn, e1 := net.DialTCP("tcp", nil, r.config.ServerAddress)
	if e1 != nil {
		fmt.Printf("[Miner Relay Service] connect to server <%s> error:\n", r.config.ServerAddress.String())
		fmt.Println(e1)
		fmt.Println("[Miner Relay Service] Reconnection will be initiated in two minutes...")
		return
	}
	fmt.Println("success.")

	go r.handleServerConn(conn)

}

func (r *RelayService) handleServerConn(conn *net.TCPConn) {

	r.service_tcp = conn
	defer func() {
		r.service_tcp = nil
	}()

	// 开始处理消息

	// 已连接上
	respmsgobj, err := message.HandleConnectToServer(conn, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 是否接受算力统计
	if respmsgobj.AcceptHashrateStatistics.Is(false) {
		//m.config.IsReportHashrate = false // 不接受统计
		//m.powWorker.CloseUploadHashrate()    // 关闭统计
		fmt.Println("note: pool is not accept PoW power statistics.")
	}

	// 循环收取挖矿消息
	for {

		//fmt.Println("循环收取挖矿消息")
		msgty, msgbody, err := message.MsgReadFromTcpConn(conn, 0)
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				// 服务器关闭
				fmt.Println("\n[Miner Relay Service] WARNING: Server close the tcp connect, reconnection will be initiated in two minutes...")
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
			// 挖矿 stuff
			r.updateNewBlockStuff(stuff)

			// 通知全部的客户端，新区块到来
			fmt.Printf("receive new block <%d> mining stuff forward to [%d] clients at time %s.\n",
				stuff.BlockHeadMeta.GetHeight(),
				len(r.allconns),
				time.Now().Format("01/02 15:04:05"),
			)
			go r.notifyAllClientNewBlockStuffByMsgBytes(msgbody)

		} else {
			fmt.Printf("message type [%d] not supported\n", msgty)
			continue
		}
	}

}
