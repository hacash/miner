package minerrelayservice

import (
	"fmt"
	interfaces2 "github.com/hacash/miner/interfaces"
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

	// Start processing message

	// Connected
	respmsgobj, err := message.HandleConnectToServer(conn, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Whether to accept calculation force statistics
	if respmsgobj.AcceptHashrateStatistics.Is(false) {
		r.config.IsReportHashrate = false // not report hashrate
		fmt.Println("note: server is not accept PoW power statistics.")
	}

	// Collect mining messages circularly
	for {

		//fmt.Println("循环收取挖矿消息")
		msgty, msgbody, err := message.MsgReadFromTcpConn(conn, 0)
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				// Server shutdown
				fmt.Println(err.Error())
				fmt.Println("\n[Miner Relay Service] WARNING: Server close the tcp connect, reconnection will be initiated in two minutes...")
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

			if r.hashratepool != nil {
				go r.hashratepool.NewMiningStuff(stuff) // notify
			}

			// cache Mining stuff
			r.updateNewBlockStuff(stuff)

			// Notify all clients of the arrival of new blocks
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
