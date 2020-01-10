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

// 监听端口
func (p *MinerPool) startServerListen() error {
	port := p.Config.TcpListenPort
	listen := net.TCPAddr{IP: net.IPv4zero, Port: port, Zone: ""}
	server, err := net.ListenTCP("tcp", &listen)
	if err != nil {
		return err
	}

	fmt.Printf("[Miner Pool] Start server and listen on port %d.\n", port)

	go func() {
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

	atomic.AddInt32(&p.currentTcpConnectingCount, 1)
	defer func() {
		atomic.AddInt32(&p.currentTcpConnectingCount, -1) // 减法
		conn.Close()
	}()

	if p.currentTcpConnectingCount > int32(p.Config.TcpConnectMaxSize) {
		conn.Write([]byte("too_many_connect"))
		return
	}

	// 如果还没有挖区块，则返回关闭，隔一段时间再次连接
	if p.currentRealtimePeriod == nil {
		conn.Write([]byte("not_ready_yet"))
		return
	}
	curperiod := p.currentRealtimePeriod
	// create client
	client := NewClient(nil, conn, curperiod.targetBlock)

	go func() {
		<-time.Tick(time.Second * 17)
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

		readbuf.Write( segdata[0:rn] )

		newbytes := readbuf.Bytes()


		//fmt.Println("MinerPool: rn, err := conn.Read(segdata)", segdata[0:rn])

		if len(newbytes) == 21 { // post address

			client.address = fields.Address(newbytes[0:21])
			// fmt.Println( client.address.ToReadable() )
			account := p.loadAccountAndAddPeriodByAddress(client.address)
			//fmt.Println("account.activeClients.Add(client) // add")
			account.activeClients.Add(client) // add
			client.belongAccount = account    // set belong
			// send mining stuff
			client.belongAccount.realtimePeriod.sendMiningStuffMsg(client.conn)

		} else if len(newbytes) == message.PowMasterMsgSize && client.belongAccount != nil {

			//fmt.Println( "message.PowMasterMsgSize", segdata[0:rn] )

			powresult := message.NewPowMasterMsg()
			powresult.Parse(newbytes, 0)
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
