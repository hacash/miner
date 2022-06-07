package message

// type
const (
	PoolAndWorkerAgreementVersionNumber uint16 = 1 // Communication protocol version number
	// Mining end type
	WorkerKindOfBlank uint8 = 0
	WorkerKindOfCPU   uint8 = 1 // CPU
	WorkerKindOfGPU   uint8 = 2 // GPU graphics card mining

)

// Error response
const (
	MsgErrorRetCodeSuccess                   uint16 = 0 // OK
	MsgErrorRetCodeConnectReadSengErr        uint16 = 1 // Message sending or reading error
	MsgErrorRetCodeAgreementVersionNumberErr uint16 = 2 // Too many connections
	MsgErrorRetCodeTooManyConnects           uint16 = 3 // Too many connections
)

// msg
const (
	// 0 is an error
	MinerWorkMsgTypeWorkerRegistration uint16 = 1 // Worker Registration
	MinerWorkMsgTypeServerResponse     uint16 = 2 // Server response
	MinerWorkMsgTypeMiningBlock        uint16 = 3 // Block mining information sent by the ore pool to the miners
	MinerWorkMsgTypeReportMiningResult uint16 = 4 // Mining result reporting
)
