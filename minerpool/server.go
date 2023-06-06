package minerpool

import (
	"bytes"
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/miner/message"
	"net"
	"sync/atomic"
	"time"
)

// Listening port
func (p *MinerPool) startServerListen() error {
	port := p.Conf.TcpListenPort
	listen := net.TCPAddr{IP: net.IPv4zero, Port: port, Zone: ""}
	server, err := net.ListenTCP("tcp", &listen)
	if err != nil {
		return err
	}

	fmt.Printf("[Miner Pool] Start server and listen on port %d.\n", port)

	go func() {
		defer server.Close()
		for {
			conn, err := server.AcceptTCP()
			if err != nil {
				continue
			}
			go p.acceptConn(conn)
		}
	}()

	return nil
}

func (p *MinerPool) acceptConn(conn *net.TCPConn) {

	defer conn.Close()

	if p.currentTcpConnectingCount > int32(p.Conf.TcpConnectMaxSize) {
		conn.Write([]byte("too_many_connect"))
		return
	}

	// If the block has not been excavated, it will be closed and connected again after a period of time
	if p.currentRealtimePeriod == nil || p.currentRealtimePeriod.targetBlock == nil {
		conn.Write([]byte("not_ready_yet"))
		return
	}
	// create client
	client := NewClient(nil, conn)

	atomic.AddInt32(&p.currentTcpConnectingCount, 1)
	defer atomic.AddInt32(&p.currentTcpConnectingCount, -1) // subtraction

	go func() {
		time.Sleep(time.Second * 17)
		if client.address == nil {
			conn.Close() // err end
		}
	}()

	// read msg
	segdata := make([]byte, 2048)

	for {

		readbuf := bytes.NewBuffer([]byte{})

	READNEXTBUF:

		rn, err := conn.Read(segdata)
		if err != nil {
			break
		}

		readbuf.Write(segdata[0:rn])

		newbytes := readbuf.Bytes()

		//fmt.Println("MinerPool: rn, err := conn.Read(segdata)", segdata[0:rn])

		if len(newbytes) == 4+5 && string(newbytes[0:4]) == "ping" {
			if client.belongAccount.realtimePeriod.IsOverEndBlock(newbytes[4:]) {
				conn.Write([]byte("end_current_mining")) // return end mining
			} else {
				conn.Write([]byte("pong")) // ok pong
			}

		} else if len(newbytes) == 21 { // post address

			client.address = fields.Address(newbytes[0:21])
			// fmt.Println( client.address.ToReadable() )
			account := p.loadAccountAndAddPeriodByAddress(client.address)
			//fmt.Println("account.activeClients.Add(client) // add")
			account.activeClients.Add(client) // add
			client.belongAccount = account    // set belong
			// send mining stuff
			client.belongAccount.realtimePeriod.sendMiningStuffMsg(client)

		} else if len(newbytes) == message.PowMasterMsgSize && client.belongAccount != nil {

			//fmt.Println( "message.PowMasterMsgSize", segdata[0:rn] )

			powresult := message.NewPowMasterMsg()
			_, e := powresult.Parse(newbytes, 0)
			if e != nil {
				// Parsing error, do nothing
				continue
			}
			client.postPowResult(powresult) // return pow results

		} else {

			goto READNEXTBUF

		}

	}

	// end
	//fmt.Println("client.belongAccount.activeClients.Remove(client)")
	if client.belongAccount != nil {
		client.belongAccount.activeClients.Remove(client) // remove
	}

}
