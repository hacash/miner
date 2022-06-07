package message

import (
	"fmt"
	"github.com/hacash/core/fields"
	"net"
	"time"
)

// Connect to server
func HandleConnectToServer(conn *net.TCPConn, rewardaddr *fields.Address) (*MsgServerResponse, error) {

	// Connected
	if rewardaddr == nil {
		rewardaddr, _ = fields.CheckReadableAddress("1AVRuFXNFi3rdMrPH4hdqSgFrEBnWisWaS")
	}

	// register
	var regmsgobj = MsgWorkerRegistration{
		fields.VarUint2(PoolAndWorkerAgreementVersionNumber),
		fields.VarUint1(WorkerKindOfBlank),
		*rewardaddr,
	}
	// Send registration message
	err := MsgSendToTcpConn(conn, MinerWorkMsgTypeWorkerRegistration, regmsgobj.Serialize())
	if err != nil {
		return nil, err
	}

	// Read response
	//fmt.Println("读取响应")
	msgty, msgbody, err := MsgReadFromTcpConn(conn, MsgWorkerServerResponseSize)
	if err != nil {
		return nil, err
	}
	if msgty != MinerWorkMsgTypeServerResponse {
		return nil, fmt.Errorf("respone from is not MinerWorkMsgTypeServerResponse")

	}

	// Response message
	var respmsgobj = MsgServerResponse{}
	_, err = respmsgobj.Parse(msgbody, 0)
	if err != nil {
		return nil, fmt.Errorf("message.MsgServerResponse.Parse Error", err)

	}

	if respmsgobj.RetCode != 0 {
		return nil, fmt.Errorf("ServerResponse RetCode Error", respmsgobj.RetCode)
	}

	// Communication successful
	return &respmsgobj, nil
}

// Server response
func SendServerResponseByRetCode(conn *net.TCPConn, retcode uint16) {

	//fmt.Println("sendServerResponseByRetCode", retcode)
	// Response message
	var sevresMsg = MsgServerResponse{
		RetCode:                  fields.VarUint2(retcode),
		AcceptHashrateStatistics: 0, // 不接受算力统计
	}
	SendServerResponse(conn, &sevresMsg)
}

// Server response
func SendServerResponse(conn *net.TCPConn, resmsg *MsgServerResponse) {

	//fmt.Println("sendServerResponseByRetCode", retcode)
	// Response message
	msgbodybts := resmsg.Serialize()
	err := MsgSendToTcpConn(conn, MinerWorkMsgTypeServerResponse, msgbodybts)
	if err != nil {
		fmt.Println(err)
	}
}

// Connect to client
func HandleConnectToClient(conn *net.TCPConn, isAcceptHashrateStatistics bool) (*MsgWorkerRegistration, error) {

	var isreged bool = false // Whether to complete the registration
	go func() {
		time.Sleep(time.Second * 10)
		if isreged == false {
			//fmt.Println("关闭超时未注册的连接")
			conn.Close() // Close unregistered connections for timeout
		}
	}()

	//fmt.Println("1111")
	// read msg
	ty, body, e0 := MsgReadFromTcpConn(conn, MsgWorkerRegistrationSize)
	if e0 != nil {
		//fmt.Println("e0", e0)
		SendServerResponseByRetCode(conn, MsgErrorRetCodeConnectReadSengErr)
		// The first message must be submitted
		return nil, fmt.Errorf("read msg err")
	}

	//fmt.Println("2222")

	if ty != MinerWorkMsgTypeWorkerRegistration {
		//fmt.Printf("ty != message.MinerWorkMsgTypeWorkerRegistration, = %d\n", ty)
		return nil, fmt.Errorf("msg type err")
	}
	var workerReg = MsgWorkerRegistration{}
	_, e1 := workerReg.Parse(body, 0)
	if e1 != nil {
		//fmt.Println("Parse error", e0)
		SendServerResponseByRetCode(conn, MsgErrorRetCodeConnectReadSengErr)
		// Parsing message error
		return nil, fmt.Errorf("msg parse err")
	}

	//fmt.Println("3333")

	// Check version
	if uint16(workerReg.PoolAndWorkerAgreementVersionNumber) != PoolAndWorkerAgreementVersionNumber {
		//fmt.Println("Parse error", e0)
		SendServerResponseByRetCode(conn, MsgErrorRetCodeAgreementVersionNumberErr)
		return nil, fmt.Errorf("version err")
	}

	//fmt.Println("4444")
	// Send connection success message success code
	SendServerResponse(conn, &MsgServerResponse{
		RetCode:                  fields.VarUint2(MsgErrorRetCodeSuccess),
		AcceptHashrateStatistics: fields.CreateBool(isAcceptHashrateStatistics),
	})

	isreged = true // Registration succeeded, do not close
	// success
	return &workerReg, nil
}
