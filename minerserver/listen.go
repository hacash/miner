package minerserver

import (
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/miner/message"
	"net"
	"time"
)

func (m *MinerServer) startListen() error {

	port := int(m.config.TcpListenPort)
	listen := net.TCPAddr{IP: net.IPv4zero, Port: port, Zone: ""}
	server, err := net.ListenTCP("tcp", &listen)
	if err != nil {
		return err
	}

	fmt.Printf("[Miner Server] Start server and listen on port %d.\n", port)

	go func() {
		for {
			conn, err := server.AcceptTCP()
			if err != nil {
				continue
			}
			go m.acceptConn(conn)
		}
	}()

	return nil
}

func sendServerResponseByRetCode(conn *net.TCPConn, retcode uint16) {

	//fmt.Println("sendServerResponseByRetCode", retcode)
	// 响应消息
	var sevresMsg = message.MsgServerResponse{
		RetCode:               fields.VarUint2(retcode),
		AcceptPowerStatistics: 0, // 不接受算力统计
	}
	msgbodybts := sevresMsg.Serialize()
	err := message.MsgSendToTcpConn(conn, message.MinerWorkMsgTypeServerResponse, msgbodybts)
	if err != nil {
		fmt.Println(err)
	}
}

func (m *MinerServer) acceptConn(conn *net.TCPConn) {

	defer conn.Close()

	// 连接数过多
	if len(m.allconns) > m.config.MaxWorkerConnect {
		sendServerResponseByRetCode(conn, message.MsgErrorRetCodeTooManyConnects)
		return
	}

	var isreged bool = false // 是否完成报名注册
	go func() {
		time.Sleep(time.Second * 10)
		if isreged == false {
			//fmt.Println("关闭超时未注册的连接")
			conn.Close() // 关闭超时未注册的连接
		}
	}()

	//fmt.Println("1111")
	// read msg
	ty, body, e0 := message.MsgReadFromTcpConn(conn, message.MsgWorkerRegistrationSize)
	if e0 != nil {
		//fmt.Println("e0", e0)
		sendServerResponseByRetCode(conn, message.MsgErrorRetCodeConnectReadSengErr)
		return
	}

	//fmt.Println("2222")

	if ty != message.MinerWorkMsgTypeWorkerRegistration {
		//fmt.Printf("ty != message.MinerWorkMsgTypeWorkerRegistration, = %d\n", ty)
		return // 第一条信息必须为上报
	}
	var workerReg = message.MsgWorkerRegistration{}
	_, e1 := workerReg.Parse(body, 0)
	if e1 != nil {
		//fmt.Println("Parse error", e0)
		sendServerResponseByRetCode(conn, message.MsgErrorRetCodeConnectReadSengErr)
		return // 解析消息错误
	}

	//fmt.Println("3333")

	// 检查版本
	if uint16(workerReg.PoolAndWorkerAgreementVersionNumber) != message.PoolAndWorkerAgreementVersionNumber {
		//fmt.Println("Parse error", e0)
		sendServerResponseByRetCode(conn, message.MsgErrorRetCodeAgreementVersionNumberErr)
		return
	}

	//fmt.Println("4444")
	// 发送连接成功消息 SUCCESS CODE
	sendServerResponseByRetCode(conn, message.MsgErrorRetCodeSuccess)

	//fmt.Println("5555")
	// 发送区块挖掘消息
	if m.penddingBlockMsg != nil {
		msgbody := m.penddingBlockMsg.Serialize()
		err := message.MsgSendToTcpConn(conn, message.MinerWorkMsgTypeMiningBlock, msgbody)
		if err != nil {
			//fmt.Println("MsgSendToTcpConn error", e0)
			sendServerResponseByRetCode(conn, message.MsgErrorRetCodeConnectReadSengErr)
			return // 解析消息错误
		}
	}
	//fmt.Println("6666")

	// 创建 client
	client := NewMinerServerClinet(m, conn)

	// 添加
	m.allconns[client.id] = client

	isreged = true // 完成上报

	//fmt.Println("+++++++++++")
	client.Handle()
	//fmt.Println("-----------")

	// 失败或关闭
	delete(m.allconns, client.id)

}
