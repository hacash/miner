package message

import (
	"fmt"
	"github.com/hacash/core/fields"
	"net"
	"time"
)

// 连接到服务器
func HandleConnectToServer(conn *net.TCPConn, rewardaddr *fields.Address) (*MsgServerResponse, error) {

	// 已连接上
	if rewardaddr == nil {
		rewardaddr, _ = fields.CheckReadableAddress("1AVRuFXNFi3rdMrPH4hdqSgFrEBnWisWaS")
	}

	// 注册
	var regmsgobj = MsgWorkerRegistration{
		fields.VarUint2(PoolAndWorkerAgreementVersionNumber),
		fields.VarUint1(WorkerKindOfBlank),
		*rewardaddr,
	}
	// 发送注册消息
	err := MsgSendToTcpConn(conn, MinerWorkMsgTypeWorkerRegistration, regmsgobj.Serialize())
	if err != nil {
		return nil, err
	}

	// 读取响应
	//fmt.Println("读取响应")
	msgty, msgbody, err := MsgReadFromTcpConn(conn, MsgWorkerServerResponseSize)
	if err != nil {
		return nil, err
	}
	if msgty != MinerWorkMsgTypeServerResponse {
		return nil, fmt.Errorf("respone from is not MinerWorkMsgTypeServerResponse")

	}

	// 响应消息
	var respmsgobj = MsgServerResponse{}
	_, err = respmsgobj.Parse(msgbody, 0)
	if err != nil {
		return nil, fmt.Errorf("message.MsgServerResponse.Parse Error", err)

	}

	if respmsgobj.RetCode != 0 {
		return nil, fmt.Errorf("ServerResponse RetCode Error", respmsgobj.RetCode)
	}

	// 通信成功
	return &respmsgobj, nil
}

// 服务端响应
func SendServerResponseByRetCode(conn *net.TCPConn, retcode uint16) {

	//fmt.Println("sendServerResponseByRetCode", retcode)
	// 响应消息
	var sevresMsg = MsgServerResponse{
		RetCode:               fields.VarUint2(retcode),
		AcceptPowerStatistics: 0, // 不接受算力统计
	}
	msgbodybts := sevresMsg.Serialize()
	err := MsgSendToTcpConn(conn, MinerWorkMsgTypeServerResponse, msgbodybts)
	if err != nil {
		fmt.Println(err)
	}
}

// 连接到客户端
func HandleConnectToClient(conn *net.TCPConn) (*MsgWorkerRegistration, error) {

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
	ty, body, e0 := MsgReadFromTcpConn(conn, MsgWorkerRegistrationSize)
	if e0 != nil {
		//fmt.Println("e0", e0)
		SendServerResponseByRetCode(conn, MsgErrorRetCodeConnectReadSengErr)
		// 第一条消息必须为上报
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
		// 解析消息错误
		return nil, fmt.Errorf("msg parse err")
	}

	//fmt.Println("3333")

	// 检查版本
	if uint16(workerReg.PoolAndWorkerAgreementVersionNumber) != PoolAndWorkerAgreementVersionNumber {
		//fmt.Println("Parse error", e0)
		SendServerResponseByRetCode(conn, MsgErrorRetCodeAgreementVersionNumberErr)
		return nil, fmt.Errorf("version err")
	}

	//fmt.Println("4444")
	// 发送连接成功消息 SUCCESS CODE
	SendServerResponseByRetCode(conn, MsgErrorRetCodeSuccess)

	isreged = true // 注册成功，不关闭
	// 成功
	return &workerReg, nil
}
