package message

import "github.com/hacash/core/interfacev2"

type PowDeviceWorker interface {
	Excavate(inputblockheadmeta interfacev2.Block, outputCh chan PowMasterMsg)

	SetCoinbaseMsgNum(coinbaseMsgNum uint32)

	StopMining()
}
