package miningpool

import (
	"github.com/hacash/core/fields"
	"github.com/hacash/miner/message"
	"net"
	"sync/atomic"
	"time"
)

// 监听端口
func (p *MinerPool) startServerListen() error {
	listen := net.TCPAddr{IP: net.IPv4zero, Port: p.config.TcpListenPort, Zone: ""}
	server, err := net.ListenTCP("tcp", &listen)
	if err != nil {
		return err
	}

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

	if p.currentTcpConnectingCount > uint32(p.config.TcpConnectMaxSize) {
		conn.Write([]byte("too_many_connect"))
		conn.Close() // 连接最大值
		return
	}

	atomic.AddUint32(&p.currentTcpConnectingCount, 1)
	defer func() {
		atomic.AddUint32(&p.currentTcpConnectingCount, -1) // 减法
	}()

	// 如果还没有挖区块，则返回关闭，隔一段时间再次连接
	if p.currentRealtimePeriod == nil {
		conn.Write([]byte("not_ready_yet"))
		conn.Close()
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
		rn, err := conn.Read(segdata)
		if err != nil {
			break
		}
		if rn == 21 && client.address == nil { // post address
			client.address = fields.Address(segdata[0:21])
			account := p.loadAccountAndAddPeriodByAddress(client.address)
			account.activeClients.Add(client) // add
			client.belongAccount = account    // set belong
			// send mining stuff
			client.belongAccount.realtimePeriod.sendMiningStuffMsg(client.conn)
		} else if rn == message.PowMasterMsgSize && client.belongAccount != nil {
			powresult := message.NewPowMasterMsg()
			powresult.Parse(segdata[0:rn], 0)
			client.postPowResult(powresult) // return pow results
			break                           // close conn
		}
	}

	// end
	client.belongAccount.activeClients.Remove(client) // remove
	conn.Close()

}
