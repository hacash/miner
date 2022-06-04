package minerrelayservice

import (
	"fmt"
	"github.com/hacash/miner/message"
	"net"
)

func (r *RelayService) startListen() {

	port := int(r.config.ServerTcpListenPort)
	if port == 0 {
		// Do not start the server
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

	// Too many connections
	if len(r.allconns) > r.config.MaxWorkerConnect {
		message.SendServerResponseByRetCode(conn, message.MsgErrorRetCodeTooManyConnects)
		return
	}

	regobj, err := message.HandleConnectToClient(conn, r.config.IsAcceptHashrate)
	if err != nil {
		return // Registration error
	}

	//fmt.Println("5555")
	// Send block mining message
	if r.penddingBlockStuff != nil {
		msgbody := r.penddingBlockStuff.Serialize()
		err := message.MsgSendToTcpConn(conn, message.MinerWorkMsgTypeMiningBlock, msgbody)
		if err != nil {
			//fmt.Println("MsgSendToTcpConn error", e0)
			message.SendServerResponseByRetCode(conn, message.MsgErrorRetCodeConnectReadSengErr)
			return // Parsing message error
		}
	}
	//fmt.Println("6666")

	// Create client
	client := NewConnClient(r, conn, regobj.RewardAddress)

	// add to
	r.addClient(client)

	//fmt.Println("+++++++++++")
	client.Handle()
	//fmt.Println("-----------")

	// Failed or closed
	r.dropClient(client)

}
