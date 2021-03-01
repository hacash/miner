package message

import "github.com/hacash/core/interfaces"

type PowDeviceWorker interface {
	Excavate(inputblockheadmeta interfaces.Block, outputCh chan PowMasterMsg)

	SetCoinbaseMsgNum(coinbaseMsgNum uint32)

	StopMining()
}
