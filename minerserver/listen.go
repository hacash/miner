package minerserver

import (
	"fmt"
	"github.com/hacash/miner/message"
	"net"
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

func (m *MinerServer) acceptConn(conn *net.TCPConn) {

	defer conn.Close()

	// 连接数过多
	if len(m.allconns) > m.config.MaxWorkerConnect {
		message.SendServerResponseByRetCode(conn, message.MsgErrorRetCodeTooManyConnects)
		return
	}

	_, err := message.HandleConnectToClient(conn, false)
	if err != nil {
		return // 注册错误
	}

	//fmt.Println("5555")
	// 发送区块挖掘消息
	if m.penddingBlockMsg != nil {
		msgbody := m.penddingBlockMsg.Serialize()
		err := message.MsgSendToTcpConn(conn, message.MinerWorkMsgTypeMiningBlock, msgbody)
		if err != nil {
			//fmt.Println("MsgSendToTcpConn error", e0)
			message.SendServerResponseByRetCode(conn, message.MsgErrorRetCodeConnectReadSengErr)
			return // 解析消息错误
		}
	}
	//fmt.Println("6666")

	// 创建 client
	client := NewMinerServerClinet(m, conn)

	// 添加
	m.changelock.Lock()
	m.allconns[client.id] = client
	m.changelock.Unlock()

	//fmt.Println("+++++++++++")
	client.Handle()
	//fmt.Println("-----------")

	// 失败或关闭
	m.changelock.Lock()
	delete(m.allconns, client.id)
	m.changelock.Unlock()

}
