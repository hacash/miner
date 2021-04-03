package minerrelayservice

import (
	"fmt"
	"github.com/hacash/miner/message"
	"net"
)

func (r *RelayService) startListen() {

	port := int(r.config.TcpListenPort)
	if port == 0 {
		// 不启动服务器
		fmt.Println("config server_listen_port==0 do not start server.")
		return
	}

	listen := net.TCPAddr{IP: net.IPv4zero, Port: port, Zone: ""}
	server, err := net.ListenTCP("tcp", &listen)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("[Miner Relay Service] Start server and listen on port %d.\n", port)

	go func() {
		for {
			conn, err := server.AcceptTCP()
			if err != nil {
				continue
			}
			go r.acceptConn(conn)
		}
	}()
}

func (r *RelayService) acceptConn(conn *net.TCPConn) {

	defer conn.Close()

	// 连接数过多
	if len(r.allconns) > r.config.MaxWorkerConnect {
		message.SendServerResponseByRetCode(conn, message.MsgErrorRetCodeTooManyConnects)
		return
	}

	_, err := message.HandleConnectToClient(conn, r.config.IsAcceptHashrate)
	if err != nil {
		return // 注册错误
	}

	//fmt.Println("5555")
	// 发送区块挖掘消息
	if r.penddingBlockStuff != nil {
		msgbody := r.penddingBlockStuff.Serialize()
		err := message.MsgSendToTcpConn(conn, message.MinerWorkMsgTypeMiningBlock, msgbody)
		if err != nil {
			//fmt.Println("MsgSendToTcpConn error", e0)
			message.SendServerResponseByRetCode(conn, message.MsgErrorRetCodeConnectReadSengErr)
			return // 解析消息错误
		}
	}
	//fmt.Println("6666")

	// 创建 client
	client := NewConnClient(r, conn)

	// 添加
	r.addClient(client)

	//fmt.Println("+++++++++++")
	client.Handle()
	//fmt.Println("-----------")

	// 失败或关闭
	r.dropClient(client)

}
