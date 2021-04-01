package minerworker

import (
	"fmt"
	"github.com/hacash/core/fields"
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
	defer func() {
		m.powWorker.DoNextMining(0) // 关闭挖矿
		m.conn = nil                // 表示断开连接
	}()

	// 已连接上
	// 注册
	var regmsgobj = message.MsgWorkerRegistration{
		fields.VarUint2(message.PoolAndWorkerAgreementVersionNumber),
		fields.VarUint1(message.WorkerKindOfBlank),
		m.config.Rewards,
	}
	// 发送注册消息
	err := message.MsgSendToTcpConn(conn, message.MinerWorkMsgTypeWorkerRegistration, regmsgobj.Serialize())
	if err != nil {
		fmt.Println(err)
		return
	}

	// 读取响应
	//fmt.Println("读取响应")
	msgty, msgbody, err := message.MsgReadFromTcpConn(conn, message.MsgWorkerServerResponseSize)
	if err != nil {
		fmt.Println(err)
		return
	}
	if msgty != message.MinerWorkMsgTypeServerResponse {
		fmt.Printf("respone from %s is not MinerWorkMsgTypeServerResponse", m.config.PoolAddress.String())
		return
	}

	// 响应消息
	var respmsgobj = message.MsgServerResponse{}
	_, err = respmsgobj.Parse(msgbody, 0)
	if err != nil {
		fmt.Println("message.MsgServerResponse.Parse Error", err)
		return
	}

	if respmsgobj.RetCode != 0 {
		fmt.Println("ServerResponse RetCode Error", respmsgobj.RetCode)
		return
	}

	// 是否接受算力统计
	if respmsgobj.AcceptPowerStatistics.Is(false) {
		m.config.IsReportPower = false // 不接受统计
		m.powWorker.CloseUploadPower() // 关闭统计
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
