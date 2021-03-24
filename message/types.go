package message

// type
const (
	PoolAndWorkerAgreementVersionNumber uint16 = 1 // 通信协议版本号
	// 挖矿端类型
	WorkerKindOfBlank uint8 = 0
	WorkerKindOfCPU   uint8 = 1 // CPU
	WorkerKindOfGPU   uint8 = 2 // GPU 显卡挖矿

)

// 错误响应
const (
	MsgErrorRetCodeSuccess                   uint16 = 0 // OK
	MsgErrorRetCodeConnectReadSengErr        uint16 = 1 // 消息发送或读取错误
	MsgErrorRetCodeAgreementVersionNumberErr uint16 = 2 // 连接数太多
	MsgErrorRetCodeTooManyConnects           uint16 = 3 // 连接数太多
)

// msg
const (
	//  0 为错误
	MinerWorkMsgTypeWorkerRegistration uint16 = 1 // worker 注册
	MinerWorkMsgTypeServerResponse     uint16 = 2 // server 响应
	MinerWorkMsgTypeMiningBlock        uint16 = 3 // 矿池发给矿工的区块挖矿信息
	MinerWorkMsgTypeReportMiningResult uint16 = 4 // 挖掘结果上报
)
