package minerserver

import (
	"fmt"
	"github.com/hacash/miner/message"
	"net"
)

func (m *MinerServer) startListen() error {

	port := int(m.Conf.TcpListenPort)
	listen := net.TCPAddr{IP: net.IPv4zero, Port: port, Zone: ""}
	server, err := net.ListenTCP("tcp", &listen)
	if err != nil {
		return err
	}

	fmt.Printf("[Miner Server] Start server and listen on port %d.\n", port)

	go func() {
		defer server.Close()
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

	// Too many connections
	if len(m.allconns) > m.Conf.MaxWorkerConnect {
		message.SendServerResponseByRetCode(conn, message.MsgErrorRetCodeTooManyConnects)
		return
	}

	_, err := message.HandleConnectToClient(conn, false)
	if err != nil {
		return // Registration error
	}

	//fmt.Println("5555")
	// Send block mining message
	if m.penddingBlockMsg != nil {
		msgbody, err := m.penddingBlockMsg.Serialize()
		err = message.MsgSendToTcpConn(conn, message.MinerWorkMsgTypeMiningBlock, msgbody)
		if err != nil {
			//fmt.Println("MsgSendToTcpConn error", e0)
			message.SendServerResponseByRetCode(conn, message.MsgErrorRetCodeConnectReadSengErr)
			return // Parsing message error
		}
	}
	//fmt.Println("6666")

	// Create client
	client := NewMinerServerClinet(m, conn)

	// add to
	m.changelock.Lock()
	m.allconns[client.id] = client
	m.changelock.Unlock()

	//fmt.Println("+++++++++++")
	client.Handle()
	//fmt.Println("-----------")

	// Failed or closed
	m.changelock.Lock()
	delete(m.allconns, client.id)
	m.changelock.Unlock()

}
